import { h, text, type VNode, type Action } from "hyperapp";
import type { State, HabitWithLogs, HabitLog } from "../types";
import { LogHabitFx, UpdateHabitFx, ToggleHabitLogFx } from "../state";
import { toYYYYMMDD, getDatesForYear, groupLogsByYear } from "../lib/date";
import { Button } from "../components/Button";

type HSL = { hue: number; saturation: number; lightness: number };
type Color = { background: HSL; monthBackground: HSL; cellBorder: HSL; cell: HSL; text: HSL };
const HSLToString = (color: HSL): string => `hsl(${color.hue} ${color.saturation}% ${color.lightness}%)`;
const getColor = (hue: number): Color => ({
  background: { hue, saturation: 54, lightness: 13 },
  monthBackground: { hue, saturation: 60, lightness: 19 },
  cellBorder: { hue, saturation: 54, lightness: 29 },
  cell: { hue, saturation: 100, lightness: 60 },
  text: { hue, saturation: 60, lightness: 50 },
});

// NOTE: all of following magic calculations are LLM generated and I have absolutely no idea how do they work, but at least they works :)
// TODO: could I make following code more simple and readble?

const CELL_SIZE = 12;
const CELL_GAP = 6;
const BLOB_PADDING = 1;
const BLOB_RADIUS = 6;
const MONTH_LABEL_GUTTER = 25;
const STEP = CELL_SIZE + CELL_GAP;

type Point = { x: number; y: number };

const segmentKey = (p1: Point, p2: Point): string => {
  if (p1.x < p2.x || (p1.x === p2.x && p1.y < p2.y)) {
    return `${p1.x},${p1.y}:${p2.x},${p2.y}`;
  }
  return `${p2.x},${p2.y}:${p1.x},${p1.y}`;
};

const generateOutline = (coords: { w: number; d: number }[]): Point[] => {
  if (coords.length === 0) return [];
  const segments = new Map<string, number>();
  coords.forEach(({ w, d }) => {
    const p00 = { x: w, y: d }; const p10 = { x: w + 1, y: d };
    const p01 = { x: w, y: d + 1 }; const p11 = { x: w + 1, y: d + 1 };
    [segmentKey(p00, p10), segmentKey(p10, p11), segmentKey(p01, p11), segmentKey(p00, p01)].forEach(key => {
      segments.set(key, (segments.get(key) || 0) + 1);
    });
  });
  const outlineSegments: { p1: Point; p2: Point }[] = []; const pointStrings = new Set<string>();
  segments.forEach((count, key) => {
    if (count === 1) {
      const [p1Str, p2Str] = key.split(':');
      const [x1, y1] = p1Str.split(',').map(Number); const [x2, y2] = p2Str.split(',').map(Number);
      outlineSegments.push({ p1: { x: x1, y: y1 }, p2: { x: x2, y: y2 } });
      pointStrings.add(p1Str); pointStrings.add(p2Str);
    }
  });
  if (outlineSegments.length === 0) return [];
  const adj = new Map<string, Point[]>(); pointStrings.forEach(pStr => adj.set(pStr, []));
  outlineSegments.forEach(({ p1, p2 }) => {
    const p1Str = `${p1.x},${p1.y}`; const p2Str = `${p2.x},${p2.y}`;
    adj.get(p1Str)!.push(p2); adj.get(p2Str)!.push(p1);
  });
  let currentPoint = outlineSegments[0].p1; const path: Point[] = []; const visited = new Set<string>();
  while (path.length < pointStrings.size) {
    const currentPointStr = `${currentPoint.x},${currentPoint.y}`;
    path.push(currentPoint); visited.add(currentPointStr);
    const neighbors = adj.get(currentPointStr)!;
    const next = neighbors.find(n => !visited.has(`${n.x},${n.y}`));
    if (!next) break;
    currentPoint = next;
  }
  const corners: Point[] = [];
  if (path.length > 2) {
    for (let i = 0; i < path.length; i++) {
      const p_prev = path[(i + path.length - 1) % path.length]; const p_curr = path[i]; const p_next = path[(i + 1) % path.length];
      const v_in = { x: p_curr.x - p_prev.x, y: p_curr.y - p_prev.y }; const v_out = { x: p_next.x - p_curr.x, y: p_next.y - p_curr.y };
      if (v_in.x * v_out.y - v_in.y * v_out.x !== 0) { corners.push(p_curr); }
    }
  } else { return path; }
  if (corners.length < 3) return path;
  let area = 0;
  for (let i = 0; i < corners.length; i++) { const p1 = corners[i]; const p2 = corners[(i + 1) % corners.length]; area += (p1.x * p2.y - p2.x * p1.y); }
  if (area < 0) { corners.reverse(); }
  return corners;
};

const createBlobPath = (polygon: Point[], options: { padding: number; radius: number, offsetX: number, offsetY: number }): string => {
  const { padding, radius, offsetX, offsetY } = options; const path = []; const len = polygon.length; if (len < 3) return "";
  const getPixelPos = (p: Point) => ({ x: p.x * STEP - CELL_GAP / 2 - offsetX, y: p.y * STEP - CELL_GAP / 2 - offsetY });
  const pathPoints: { p_before: Point; p_after: Point; sweepFlag: number; radius: number }[] = [];
  for (let i = 0; i < len; i++) {
    const p_prev = polygon[(i + len - 1) % len]; const p_curr = polygon[i]; const p_next = polygon[(i + 1) % len];
    const pos_curr = getPixelPos(p_curr); const pos_prev = getPixelPos(p_prev); const pos_next = getPixelPos(p_next);
    const v_in = { x: pos_curr.x - pos_prev.x, y: pos_curr.y - pos_prev.y }; const v_out = { x: pos_next.x - pos_curr.x, y: pos_next.y - pos_curr.y };
    const len_in = Math.sqrt(v_in.x ** 2 + v_in.y ** 2); const v_in_n = { x: v_in.x / len_in, y: v_in.y / len_in };
    const len_out = Math.sqrt(v_out.x ** 2 + v_out.y ** 2); const v_out_n = { x: v_out.x / len_out, y: v_out.y / len_out };
    const v_in_p = { x: -v_in_n.y, y: v_in_n.x }; const v_out_p = { x: -v_out_n.y, y: v_out_n.x };
    const cross_z = v_in_n.x * v_out_n.y - v_in_n.y * v_out_n.x; const isOuter = cross_z > 0; const sweepFlag = isOuter ? 1 : 0;
    const cornerRadius = isOuter ? radius : radius + BLOB_PADDING * 2;
    const t_before = { x: pos_curr.x - v_in_n.x * cornerRadius, y: pos_curr.y - v_in_n.y * cornerRadius };
    const t_after = { x: pos_curr.x + v_out_n.x * cornerRadius, y: pos_curr.y + v_out_n.y * cornerRadius };
    const p_before = { x: t_before.x + v_in_p.x * padding, y: t_before.y + v_in_p.y * padding };
    const p_after = { x: t_after.x + v_out_p.x * padding, y: t_after.y + v_out_p.y * padding };
    pathPoints.push({ p_before, p_after, sweepFlag, radius: cornerRadius });
  }
  for (let i = 0; i < len; i++) {
    const current = pathPoints[i]; const prev = pathPoints[(i + len - 1) % len];
    if (i === 0) { path.push(`M ${prev.p_after.x} ${prev.p_after.y}`); }
    path.push(`L ${current.p_before.x} ${current.p_before.y}`);
    path.push(`A ${current.radius} ${current.radius} 0 0 ${current.sweepFlag} ${current.p_after.x} ${current.p_after.y}`);
  }
  path.push("Z"); return path.join(" ");
};

type GridCell = { date: Date; w: number; d: number };
type MonthLayout = {
  monthIndex: number;
  label: { text: string; x: number };
  svg: { x: number; y: number; width: number; height: number; path: string };
  cells: GridCell[];
};
type YearLayout = {
  year: number;
  months: MonthLayout[];
  totalWidth: number;
  totalHeight: number;
};

const generateCalendarLayout = (year: number): YearLayout => {
  const dates = getDatesForYear(year);
  if (dates.length === 0) return { year, months: [], totalWidth: 0, totalHeight: 0 };

  const yearStart = new Date(Date.UTC(year, 0, 1));
  const firstDayOfWeek = yearStart.getUTCDay();

  const getGridPos = (date: Date) => {
    const dayOfYear = Math.floor((date.getTime() - yearStart.getTime()) / 86400000);
    const dayOfWeek = date.getUTCDay();
    const weekOfYear = Math.floor((dayOfYear + firstDayOfWeek) / 7);
    return { w: weekOfYear, d: dayOfWeek };
  };

  const allCells = dates.map(date => ({ date, ...getGridPos(date) }));
  const monthlyCells: GridCell[][] = Array.from({ length: 12 }, () => []);
  allCells.forEach(cell => monthlyCells[cell.date.getUTCMonth()].push(cell));

  const months: MonthLayout[] = monthlyCells.map((cells, monthIndex) => {
    if (cells.length === 0) return null;

    const firstDayCell = cells[0];
    const weeks = cells.map(c => c.w);
    const minWeek = Math.min(...weeks);
    const maxWeek = Math.max(...weeks);

    const cellsByWeek = new Map<number, GridCell[]>();
    for (const cell of cells) {
      if (!cellsByWeek.has(cell.w)) {
        cellsByWeek.set(cell.w, []);
      }
      cellsByWeek.get(cell.w)!.push(cell);
    }

    const sortedWeeks = Array.from(cellsByWeek.keys()).sort((a, b) => a - b);
    let firstCompleteWeekIndex: number | undefined;
    for (const weekIndex of sortedWeeks) {
      if (cellsByWeek.get(weekIndex)!.length === 7) {
        firstCompleteWeekIndex = weekIndex;
        break;
      }
    }

    const labelWeek = firstCompleteWeekIndex !== undefined
      ? firstCompleteWeekIndex
      : firstDayCell.w;

    const outline = generateOutline(cells.map(({ w, d }) => ({ w, d })));
    const svgX = minWeek * STEP - BLOB_PADDING - CELL_GAP / 2;
    const svgY = -BLOB_PADDING - CELL_GAP / 2;
    const svgWidth = (maxWeek - minWeek + 1) * STEP + 2 * BLOB_PADDING + CELL_GAP;
    const svgHeight = 7 * STEP + 2 * BLOB_PADDING + CELL_GAP;
    const pathData = createBlobPath(outline, { padding: BLOB_PADDING, radius: BLOB_RADIUS, offsetX: svgX, offsetY: svgY });

    return {
      monthIndex,
      cells,
      label: {
        text: new Date(Date.UTC(year, monthIndex)).toLocaleString('en-US', { month: 'short' }),
        x: labelWeek * STEP,
      },
      svg: { x: svgX, y: svgY, width: svgWidth, height: svgHeight, path: pathData },
    };
  }).filter((m): m is MonthLayout => m !== null);

  const maxWeek = allCells.length > 0 ? getGridPos(allCells[allCells.length - 1].date).w : 0;
  const totalWidth = (maxWeek + 1) * STEP;
  const totalHeight = 7 * STEP + MONTH_LABEL_GUTTER;

  return { year, months, totalWidth, totalHeight };
};

const CalendarGrid = ({ year, logs, isEditing, habitId, token, color }: { year: number, logs: HabitLog[], isEditing: boolean, habitId: string, token: string | null, color: Color }): VNode<State> | null => {
  const logsMap = new Map(logs.map(log => [toYYYYMMDD(new Date(log.date)), log.status]));
  const layout = generateCalendarLayout(year);
  if (layout.months.length === 0) return null;

  return h<State>("div", {
    class: "relative flex-shrink-0",
    style: { width: `${layout.totalWidth}px`, height: `${layout.totalHeight}px` }
  }, [
    ...layout.months.map(month => h<State>("svg", {
      key: `blob-${month.monthIndex}`, class: "absolute",
      style: { left: `${month.svg.x}px`, top: `${month.svg.y + MONTH_LABEL_GUTTER}px`, width: `${month.svg.width}px`, height: `${month.svg.height}px`, overflow: "visible" }
    }, h<State>("path", { d: month.svg.path, fill: HSLToString(color.monthBackground) }))),

    ...layout.months.flatMap(month => month.cells.map(cell => {
      const { w, d, date } = cell;
      const dateStr = toYYYYMMDD(date);
      const isLogged = logsMap.get(dateStr) === true;
      const canToggle = isEditing && new Date(new Date().setUTCHours(0, 0, 0, 0)) >= date;
      return h("div", {
        key: dateStr, class: `absolute rounded-sm ${canToggle ? 'cursor-pointer hover:opacity-75' : ''}`,
        style: {
          left: `${w * STEP}px`, top: `${d * STEP + MONTH_LABEL_GUTTER}px`,
          width: `${CELL_SIZE}px`, height: `${CELL_SIZE}px`,
          backgroundColor: isLogged ? HSLToString(color.cell) : "transparent",
          border: isLogged ? "none" : `1px solid ${HSLToString(color.cellBorder)}`,
        },
        title: `${date.toDateString()}: ${isLogged ? 'Completed' : 'Not completed'}`,
        onclick: (state: State) => {
          if (!canToggle || !token) return state;
          return [state, [ToggleHabitLogFx, { habitId, date: dateStr, currentStatus: isLogged, token }]];
        }
      });
    })),

    ...layout.months.map(month => h<State>("div", {
      key: `label-${month.monthIndex}`, class: "absolute text-xs",
      style: { color: HSLToString(color.text), left: `${month.label.x}px`, top: '0px' }
    }, text(month.label.text))),
  ]);
};


export const HabitCard = (habit: HabitWithLogs, isOwner: boolean, token: string | null, isEditing: boolean): VNode<State> => {
  const color = getColor(habit.colorHue);

  const wasLoggedToday = habit.logs.some(log =>
    new Date(log.date).setUTCHours(0, 0, 0, 0) === new Date().setUTCHours(0, 0, 0, 0)
  );

  const logsByYear = groupLogsByYear(habit.logs);
  const years = Object.keys(logsByYear).map(Number).sort((a, b) => b - a);

  if (years.length === 0) {
    const currentYear = new Date().getUTCFullYear();
    years.push(currentYear);
    logsByYear[currentYear] = [];
  }

  const onNameChange: Action<State, Event> = (state, event) => {
    const target = event.target as HTMLInputElement;
    const newName = target.value.trim();
    if (newName && newName !== habit.name && token) {
      return [state, [UpdateHabitFx, { habitId: habit.id, name: newName, colorHue: habit.colorHue, token }]];
    }
    target.value = habit.name;
    return state;
  };

  const onColorHueChange: Action<State, Event> = (state, event) => {
    const target = event.target as HTMLInputElement;
    const newColorHue = parseInt(target.value, 10);
    if (!isNaN(newColorHue) && newColorHue >= 0 && newColorHue <= 360 && newColorHue !== habit.colorHue && token) {
      return [state, [UpdateHabitFx, { habitId: habit.id, name: habit.name, colorHue: newColorHue, token }]];
    }
    target.value = String(habit.colorHue);
    return state;
  };

  return h<State>("div", {
    class: "py-8 rounded-4xl flex flex-col gap-8 overflow-hidden w-full",
    key: habit.id,
    style: { backgroundColor: HSLToString(color.background) }
  }, [
    h<State>("div", { class: "flex justify-between items-center px-8" }, [
      isEditing && isOwner
        ? h<State>("div", { class: "flex flex-col gap-2 w-full" }, [
          h("input", {
            class: "text-xl font-bold py-2 bg-transparent focus:outline-none w-full",
            style: { borderBottom: `1px solid ${HSLToString(color.cellBorder)}` },
            value: habit.name,
            onblur: onNameChange,
            onkeydown: (state: State, event: KeyboardEvent) => {
              if (event.key === "Enter") (event.target as HTMLInputElement).blur();
              return state;
            },
          }),
          h<State>("div", { class: "flex items-center gap-2" }, [
            h<State>("label", { for: `color-hue-${habit.id}`, class: "text-sm", style: { color: HSLToString(color.text) } }, text("Hue (0-360):")),
            h("input", {
              id: `color-hue-${habit.id}`, type: "number", min: "0", max: "360", value: habit.colorHue,
              class: "bg-transparent w-20 text-center py-2 focus:outline-none",
              style: { borderBottom: `1px solid ${HSLToString(color.cellBorder)}` },
              onblur: onColorHueChange,
              onkeydown: (state: State, event: KeyboardEvent) => {
                if (event.key === "Enter") (event.target as HTMLInputElement).blur();
                return state;
              },
            }),
          ])
        ])
        : h<State>("h3", { class: "text-xl font-bold" }, text(habit.name)),
      isOwner && !isEditing && token
        ? Button({
          disabled: wasLoggedToday,
          onclick: (state: State) => [state, [LogHabitFx, { habitId: habit.id, token: state.token! }]],
          class: "text-xs",
          style: { backgroundColor: HSLToString(color.text) }
        }, text(wasLoggedToday ? "Completed!" : "Log Today"))
        : null
    ]),

    h<State>("div", { class: "relative" }, [
      h<State>("div", { class: "flex flex-nowrap flex-row-reverse w-full overflow-x-scroll no-scrollbar items-end gap-4" },
        years.map(year =>
          h<State>("div", { key: String(year), class: "flex flex-col items-center gap-2 flex-shrink-0 px-8" }, [
            h<State>("div", { class: "font-bold", style: { color: HSLToString(color.text) } }, text(year)),
            CalendarGrid({ year, logs: logsByYear[year] || [], isEditing, habitId: habit.id, token, color }),
          ])
        )
      ),
      h<State>("div", {
        class: "absolute inset-y-0 left-0 w-8 pointer-events-none",
        style: { background: `linear-gradient(to right, ${HSLToString(color.background)}, hsla(${color.background.hue} ${color.background.saturation}% ${color.background.lightness}% / 0))` }
      }),
      h<State>("div", {
        class: "absolute inset-y-0 right-0 w-8 pointer-events-none",
        style: { background: `linear-gradient(to left, ${HSLToString(color.background)}, hsla(${color.background.hue} ${color.background.saturation}% ${color.background.lightness}% / 0))` }
      })
    ])
  ]);
};
