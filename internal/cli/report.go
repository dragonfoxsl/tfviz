package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/danushkastanley/tfviz/internal/input"
	"github.com/danushkastanley/tfviz/internal/pricing"
	"github.com/danushkastanley/tfviz/internal/report/build"
	"github.com/danushkastanley/tfviz/internal/report/html"
	"github.com/danushkastanley/tfviz/internal/report/model"
	"github.com/danushkastanley/tfviz/internal/report/safeshare"
)

type reportOptions struct {
	input       string
	output      string
	title       string
	view        string
	force       bool
	strict      bool
	offline     bool
	safeShare   bool
	icons       string
	pricingFile string
	usageFile   string
	s3          s3Flags
}

func parseReportFlags(kind input.SnapshotKind, args []string, env Env) (reportOptions, error) {
	var opts reportOptions
	fs := flag.NewFlagSet("tfviz "+string(kind), flag.ContinueOnError)
	fs.SetOutput(env.Stderr)
	noun := map[input.SnapshotKind]string{input.KindPlan: "plan", input.KindState: "state"}[kind]
	inputHelp := "exported " + noun + " JSON file, or - for standard input (required)"
	if kind == input.KindState {
		inputHelp = "state: show -json output, a version-4 .tfstate file, s3://bucket/key, or - for standard input (required)"
	}
	fs.StringVar(&opts.input, "input", "", inputHelp)
	fs.StringVar(&opts.output, "output", "", "HTML report to write (required)")
	fs.StringVar(&opts.title, "title", "", "report title")
	fs.StringVar(&opts.view, "view", "architecture", "structure the report opens in: architecture or modules")
	fs.BoolVar(&opts.force, "force", false, "replace the output file if it exists")
	fs.BoolVar(&opts.strict, "strict", false, "fail if any resource, action or input cannot be fully interpreted")
	fs.BoolVar(&opts.offline, "offline", false, "refuse any input that would need network access")
	fs.BoolVar(&opts.safeShare, "safe-share", false, "replace names, IDs, ARNs, accounts, addresses and network details with stand-ins")
	fs.StringVar(&opts.icons, "icons", "", iconsHelp)
	if kind == input.KindPlan {
		fs.StringVar(&opts.pricingFile, "pricing-file", "", "local offline user-rate catalogue for partial resource estimates")
		fs.StringVar(&opts.usageFile, "usage-file", "", "optional local monthly usage JSON (requires --pricing-file)")
	}
	if kind == input.KindState {
		fs.StringVar(&opts.s3.profile, "aws-profile", "", "AWS profile for s3:// inputs (uses the standard credential chain)")
		fs.StringVar(&opts.s3.region, "aws-region", "", "region of the state bucket for s3:// inputs")
		fs.StringVar(&opts.s3.version, "s3-version", "", "read this version of the s3:// object (needs s3:GetObjectVersion)")
		fs.StringVar(&opts.s3.owner, "expected-bucket-owner", "", "fail unless the s3:// bucket belongs to this account ID")
	}
	fs.Usage = func() {
		fmt.Fprintf(env.Stderr, "Usage: tfviz %s --input <file|-> --output <report.html> [options]\n\nOptions:\n", kind)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return opts, err
		}
		return opts, &usageError{"invalid options"}
	}
	switch {
	case opts.usageFile != "" && opts.pricingFile == "":
		return opts, &usageError{"--usage-file requires --pricing-file"}
	case fs.NArg() > 0:
		return opts, &usageError{fmt.Sprintf("unexpected argument %q", clip(fs.Arg(0)))}
	case opts.input == "":
		return opts, &usageError{"--input is required (use - for standard input)"}
	case opts.output == "":
		return opts, &usageError{"--output is required"}
	case opts.view != "architecture" && opts.view != "modules":
		return opts, &usageError{"--view must be architecture or modules"}
	case opts.output == "-":
		return opts, &usageError{"--output must be a file path; reports are not written to standard output"}
	case kind == input.KindPlan && strings.HasPrefix(opts.input, "s3://"):
		return opts, &usageError{"plans are read from a file or standard input; export the plan with show -json first"}
	}
	return opts, nil
}

func runReport(kind input.SnapshotKind, args []string, env Env) int {
	opts, err := parseReportFlags(kind, args, env)
	if errors.Is(err, flag.ErrHelp) {
		return ExitOK
	}
	if err != nil {
		fmt.Fprintf(env.Stderr, "tfviz: %v\n", err)
		return exitCode(err)
	}
	summary, err := generate(kind, opts, env)
	if err != nil {
		fmt.Fprintf(env.Stderr, "tfviz: %v\n", err)
		return exitCode(err)
	}
	fmt.Fprintln(env.Stderr, summary)
	return ExitOK
}

func generate(kind input.SnapshotKind, opts reportOptions, env Env) (string, error) {
	if err := checkOutput(opts.output, opts.force); err != nil {
		return "", err
	}
	report, err := buildReport(kind, opts, env)
	if err != nil {
		return "", err
	}
	var page bytes.Buffer
	if err := html.Render(&page, report, html.BundledAssets()); err != nil {
		return "", err
	}
	if err := writeAtomically(opts.output, page.Bytes(), opts.force); err != nil {
		return "", err
	}
	return describe(opts.output, report), nil
}

// buildReport reads the input and produces the report model, applying
// --strict and --safe-share. It is shared by report generation and explore.
func buildReport(kind input.SnapshotKind, opts reportOptions, env Env) (*model.Report, error) {
	var catalogue *pricing.Catalogue
	if opts.usageFile != "" && (opts.pricingFile == "" || kind != input.KindPlan) {
		return nil, &usageError{"--usage-file requires --pricing-file and a plan report"}
	}
	if opts.pricingFile != "" {
		if kind != input.KindPlan {
			return nil, &usageError{"--pricing-file is only supported for plan reports"}
		}
		var err error
		catalogue, err = pricing.Load(opts.pricingFile)
		if err != nil {
			return nil, err
		}
	}
	in, err := readInput(opts, env)
	if err != nil {
		return nil, err
	}
	snap, err := input.Read(in.data, kind)
	in.data = nil // the raw, secret-bearing input is not needed beyond this point
	if err != nil {
		return nil, err
	}
	if opts.usageFile != "" {
		if err := catalogue.LoadUsage(opts.usageFile, snap.Resources); err != nil {
			return nil, err
		}
	}
	report := build.Report(snap, build.Options{
		Title: opts.title, View: model.View(opts.view), Source: in.kind, SourceTime: in.time, SourceVersion: in.version,
		GeneratedAt: env.Now(), ToolVersion: env.Version,
		Pricing: catalogue,
	})
	if opts.strict {
		if problems := strictProblems(report); len(problems) > 0 {
			return nil, &usageError{"--strict: " + strings.Join(problems, "; ")}
		}
	}
	if opts.safeShare {
		report = safeshare.Apply(report)
	}
	if opts.icons != "" {
		if report.Icons, err = loadIcons(opts.icons, report); err != nil {
			return nil, err
		}
	}
	return report, nil
}

// strictProblems lists interpretation gaps that --strict turns into failure.
func strictProblems(r *model.Report) []string {
	var problems []string
	if r.Coverage.ResourcesGeneric > 0 {
		problems = append(problems, fmt.Sprintf("%d unsupported resource(s)", r.Coverage.ResourcesGeneric))
	}
	if r.Summary.Unsupported > 0 {
		problems = append(problems, fmt.Sprintf("%d unrecognised action(s)", r.Summary.Unsupported))
	}
	for _, w := range r.Warnings {
		switch w.Code {
		case model.WarningIncompletePlan, model.WarningDeposedObjectSkipped, model.WarningUnknownFormatMinor:
			problems = append(problems, w.Message)
		}
	}
	return problems
}

// describe is the one-line result: counts only, never input content.
func describe(path string, r *model.Report) string {
	s := r.Summary
	line := fmt.Sprintf("Wrote %s: %d resources (%d fully supported", path, r.Coverage.ResourcesTotal, r.Coverage.ResourcesSupported)
	if r.Coverage.ResourcesGeneric > 0 {
		line += fmt.Sprintf(", %d with limited detail", r.Coverage.ResourcesGeneric)
	}
	line += ")"
	if r.Mode == model.ModePlan {
		line += fmt.Sprintf(". Changes: %d to add, %d to change, %d to replace, %d to destroy", s.Create, s.Update, s.Replace, s.Delete)
		if s.Read > 0 {
			line += fmt.Sprintf(", %d read during apply", s.Read)
		}
		if s.Forget > 0 {
			line += fmt.Sprintf(", %d removed from state", s.Forget)
		}
	}
	if n := len(r.Unresolved); n > 0 {
		line += fmt.Sprintf(". %d unresolved reference(s)", n)
	}
	if r.Icons != nil {
		line += fmt.Sprintf(". Icons for %d resource type(s)", len(r.Icons.ByType))
	}
	return line + "."
}
