import { useCallback, useEffect, useMemo, useState } from "react";
import { ReactFlowProvider } from "@xyflow/react";
import type { Report } from "../report/schema.gen";
import type { ExploreConfig } from "../report/load";
import { indexReport, type ViewMode } from "../report/view";
import { Canvas } from "../graph/Canvas";
import { CanvasStateContext } from "../graph/canvasState";
import { ancestorGroups, selectionState } from "../graph/selection";
import { defaultCollapsed, type Structure } from "../graph/canvasModel";
import { layoutCanvas } from "../graph/layout";
import { Inspector } from "../inspector/Inspector";
import { Header } from "./Header";
import { Navigator } from "./Navigator";
import { Footer } from "./Footer";
import { CostPeriodContext, ReportIndexContext, type CostPeriod } from "./reportContext";
import { useReducedMotion, useThemePreference } from "./preferences";

export function App({ report, explore }: { report: Report; explore?: ExploreConfig | undefined }) {
  const index = useMemo(() => indexReport(report), [report]);
  const [period, setPeriod] = useState<CostPeriod>("monthly");
  const [view, setView] = useState<ViewMode>(report.mode === "plan" ? "changes" : "after");
  const [structure, setStructure] = useState<Structure>(report.initial_view ?? "architecture");
  const [collapsed, setCollapsed] = useState<ReadonlySet<string>>(() => defaultCollapsed(index, report.initial_view ?? "architecture"));
  const [selectedId, setSelectedId] = useState<string | undefined>();
  const [focusRequest, setFocusRequest] = useState<{ id: string; seq: number }>();
  const [theme, setTheme] = useThemePreference("system");
  const [reducedMotion, setReducedMotion] = useReducedMotion();

  const layout = useMemo(() => layoutCanvas(index, view, structure, collapsed), [index, view, structure, collapsed]);
  const canvasState = useMemo(() => selectionState(index, selectedId, layout.lookup), [index, selectedId, layout]);

  useEffect(() => {
    // Exposed for the performance budget checks in end-to-end tests.
    document.documentElement.dataset.layoutMs = layout.layoutMs.toFixed(1);
    document.documentElement.dataset.layoutNodes = String(layout.nodes.length);
  }, [layout]);

  const changeStructure = useCallback(
    (next: Structure) => {
      setStructure(next);
      setCollapsed(defaultCollapsed(index, next));
    },
    [index],
  );

  const toggleGroup = useCallback((groupId: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev);
      if (next.has(groupId)) next.delete(groupId);
      else next.add(groupId);
      return next;
    });
  }, []);

  // Selecting from the list expands any collapsed group on the way, then
  // brings the resource into view. Selection persists across views.
  const selectFromList = useCallback(
    (id: string) => {
      const path = ancestorGroups(index, id, structure);
      setCollapsed((prev) => (path.some((g) => prev.has(g)) ? new Set([...prev].filter((g) => !path.includes(g))) : prev));
      setSelectedId(id);
      setFocusRequest((prev) => ({ id, seq: (prev?.seq ?? 0) + 1 }));
    },
    [index, structure],
  );

  return (
    <ReportIndexContext.Provider value={index}>
      <CostPeriodContext.Provider value={{ period, setPeriod }}>
      <div className={`app${reducedMotion ? " motion-reduced" : ""}`}>
        <Header
          report={report}
          view={view}
          onViewChange={setView}
          structure={structure}
          onStructureChange={changeStructure}
          theme={theme}
          onThemeChange={setTheme}
          explore={explore}
        />
        <Navigator index={index} selectedId={selectedId} onSelect={selectFromList} />
        <main className="canvas" aria-label="Diagram">
          <ReactFlowProvider>
            <CanvasStateContext.Provider value={canvasState}>
              <Canvas
                index={index}
                layout={layout}
                layoutKey={`${structure}-${view}`}
                focusRequest={focusRequest}
                reducedMotion={reducedMotion}
                onSelect={setSelectedId}
                onToggleGroup={toggleGroup}
              />
            </CanvasStateContext.Provider>
          </ReactFlowProvider>
        </main>
        <Inspector index={index} resource={selectedId ? index.resources.get(selectedId) : undefined} onSelect={selectFromList} />
        <Footer report={report} reducedMotion={reducedMotion} onReducedMotionChange={setReducedMotion} />
      </div>
      </CostPeriodContext.Provider>
    </ReportIndexContext.Provider>
  );
}
