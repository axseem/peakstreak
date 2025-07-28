import { h, text, type VNode, type Dispatchable } from "hyperapp";
import type { State, HabitWithLogs, HabitLog, Habit, EditingHabitState } from "../types";
import {
  UpdateHabitFx,
  UpsertHabitLogFx,
  ToggleHabitMenu,
  CloseHabitMenu,
  StartEditingHabit,
  CancelEditingHabit,
  UpdateEditingHabitField,
  DeleteHabitFx,
} from "../state";
import { toYYYYMMDD, getDatesForYear, groupLogsByYear } from "../lib/date";
import { Popup } from "./Popup";
import { twMerge } from "tailwind-merge";
import { Button } from "./Button";

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
  const { padding, radius, offsetX, offsetY } = options;
  const path = [];
  const len = polygon.length;
  if (len < 3) return "";
  const getPixelPos = (p: Point) => ({ x: p.x * STEP - CELL_GAP / 2 - offsetX, y: p.y * STEP - CELL_GAP / 2 - offsetY });
  const pathPoints: { p_before: Point; p_after: Point; sweepFlag: number; radius: number }[] = [];
  for (let i = 0; i < len; i++) {
    const p_prev = polygon[(i + len - 1) % len];
    const p_curr = polygon[i]; const p_next = polygon[(i + 1) % len];
    const pos_curr = getPixelPos(p_curr);
    const pos_prev = getPixelPos(p_prev);
    const pos_next = getPixelPos(p_next);
    const v_in = { x: pos_curr.x - pos_prev.x, y: pos_curr.y - pos_prev.y };
    const v_out = { x: pos_next.x - pos_curr.x, y: pos_next.y - pos_curr.y };
    const len_in = Math.sqrt(v_in.x ** 2 + v_in.y ** 2);
    const v_in_n = { x: v_in.x / len_in, y: v_in.y / len_in };
    const len_out = Math.sqrt(v_out.x ** 2 + v_out.y ** 2);
    const v_out_n = { x: v_out.x / len_out, y: v_out.y / len_out };
    const v_in_p = { x: -v_in_n.y, y: v_in_n.x };
    const v_out_p = { x: -v_out_n.y, y: v_out_n.x };
    const cross_z = v_in_n.x * v_out_n.y - v_in_n.y * v_out_n.x;
    const isOuter = cross_z > 0;
    const sweepFlag = isOuter ? 1 : 0;
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

const CalendarGrid = ({ year, habit, logs, isEditing, token, color }: { year: number, habit: Habit, logs: HabitLog[], isEditing: boolean, token: string | null, color: Color }): VNode<State> | null => {
  const logsMap = new Map(logs.map(log => [toYYYYMMDD(new Date(log.date)), log.value]));
  const layout = generateCalendarLayout(year);
  if (layout.months.length === 0) return null;

  const maxValuesPerMonth: number[] = Array(12).fill(0);
  if (!habit.isBoolean) {
    for (const log of logs) {
      const monthIndex = new Date(log.date).getUTCMonth();
      if (log.value > maxValuesPerMonth[monthIndex]) {
        maxValuesPerMonth[monthIndex] = log.value;
      }
    }
  }

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
      const logValue = logsMap.get(dateStr) ?? 0;
      const isLogged = logValue > 0;
      const canToggle = isEditing && new Date(new Date().setUTCHours(0, 0, 0, 0)) >= date;

      let backgroundColor = "transparent";
      if (isLogged) {
        if (habit.isBoolean) {
          backgroundColor = HSLToString(color.cell);
        } else {
          const monthIndex = date.getUTCMonth();
          const maxValue = maxValuesPerMonth[monthIndex];
          // The raw opacity is the value's percentage of the max value for the month.
          const rawOpacity = maxValue > 0 ? logValue / maxValue : 0;
          // We set a minimum opacity of 10% for any logged day to ensure it's visible.
          const opacity = Math.max(0.1, rawOpacity);
          const { hue, saturation, lightness } = color.cell;
          backgroundColor = `hsla(${hue}, ${saturation}%, ${lightness}%, ${opacity})`;
        }
      }

      return h("div", {
        key: dateStr, class: `absolute rounded-sm ${canToggle ? 'cursor-pointer hover:opacity-75' : ''}`,
        style: {
          left: `${w * STEP}px`, top: `${d * STEP + MONTH_LABEL_GUTTER}px`,
          width: `${CELL_SIZE}px`, height: `${CELL_SIZE}px`,
          backgroundColor: backgroundColor,
          border: isLogged ? "none" : `1px solid ${HSLToString(color.cellBorder)}`,
        },
        title: habit.isBoolean
          ? `${date.toDateString()}: ${isLogged ? 'Completed' : 'Not completed'}`
          : `${date.toDateString()}: ${logValue}`,
        onclick: (state: State) => {
          if (!canToggle || !token) return state;

          if (habit.isBoolean) {
            const newValue = isLogged ? 0 : 1;
            return [state, [UpsertHabitLogFx, { habitId: habit.id, date: dateStr, value: newValue, token }]];
          } else {
            const rawValue = prompt(`Enter value for ${date.toDateString()}:`, String(logValue));
            if (rawValue === null) return state; // User cancelled
            const value = parseInt(rawValue, 10);
            if (!isNaN(value) && value >= 0) {
              return [state, [UpsertHabitLogFx, { habitId: habit.id, date: dateStr, value, token }]];
            } else {
              alert("Please enter a valid non-negative number.");
              return state;
            }
          }
        }
      });
    })),

    ...layout.months.map(month => h<State>("div", {
      key: `label-${month.monthIndex}`, class: "absolute text-xs",
      style: { color: HSLToString(color.text), left: `${month.label.x}px`, top: '0px' }
    }, text(month.label.text))),
  ]);
};

const HabitMenu = ({ habit, token }: { habit: HabitWithLogs, token: string | null }): VNode<State> => {
  const MenuItem = (props: { onclick: any, class?: string, confirmation?: string }, children: any) =>
    h("button", {
      class: twMerge("w-full text-left px-4 py-2 text-sm hover:bg-neutral-800 flex items-center gap-3 transition-colors rounded-md", props.class),
      onclick: (state: State) => {
        if (props.confirmation && !window.confirm(props.confirmation)) {
          return CloseHabitMenu(state);
        }
        return props.onclick(state);
      }
    }, children);

  return h("div", { class: "p-1" }, [
    MenuItem({
      onclick: (state: State) => StartEditingHabit(state, habit)
    }, [
      h("svg", { class: "w-4 h-4", fill: "none", viewBox: "0 0 24 24", "stroke-width": "1.5", stroke: "currentColor" },
        h("path", { "stroke-linecap": "round", "stroke-linejoin": "round", d: "M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" })
      ),
      text("Edit"),
    ]),
    MenuItem({
      onclick: (state: State) => [CloseHabitMenu(state), [DeleteHabitFx, { habitId: habit.id, token }]],
      class: "text-red-400 hover:bg-red-500/10",
      confirmation: `Are you sure you want to delete the habit "${habit.name}"? This action cannot be undone.`
    }, [
      h("svg", { class: "w-4 h-4", fill: "none", viewBox: "0 0 24 24", "stroke-width": "1.5", stroke: "currentColor" },
        h("path", { "stroke-linecap": "round", "stroke-linejoin": "round", d: "M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.134-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.067-2.09.92-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" })
      ),
      text("Delete")
    ]),
  ]);
};

type HabitCardProps = {
  habit: HabitWithLogs;
  isOwner: boolean;
  token: string | null;
  isEditing: boolean;
  activeHabitMenuId: string | null;
  editingHabit: EditingHabitState | null;
};

export const HabitCard = ({ habit, isOwner, token, isEditing, activeHabitMenuId, editingHabit }: HabitCardProps): VNode<State> => {
  const displayHue = isEditing && editingHabit ? editingHabit.colorHue : habit.colorHue;
  const color = getColor(displayHue);

  const todayStr = toYYYYMMDD(new Date());
  const todayLog = habit.logs.find(log => toYYYYMMDD(new Date(log.date)) === todayStr);
  const wasLoggedToday = !!todayLog && todayLog.value > 0;
  const todayValue = todayLog?.value ?? 0;

  const logsByYear = groupLogsByYear(habit.logs);
  const years = Object.keys(logsByYear).map(Number).sort((a, b) => b - a);

  if (years.length === 0) {
    const currentYear = new Date().getUTCFullYear();
    years.push(currentYear);
    logsByYear[currentYear] = [];
  }

  const handleEditSubmit = (state: State, event: Event): Dispatchable<State, any> => {
    event.preventDefault();

    if (!state.editingHabit || !token) {
      return CancelEditingHabit(state);
    }

    const newName = state.editingHabit.name.trim();
    if (newName.length === 0) {
      alert("Habit name cannot be empty.");
      return state;
    }

    const nameChanged = newName !== habit.name;
    const colorChanged = state.editingHabit.colorHue !== habit.colorHue;

    if (nameChanged || colorChanged) {
      return [
        state,
        [UpdateHabitFx, { habitId: habit.id, name: newName, colorHue: state.editingHabit.colorHue, token }]
      ];
    }

    return CancelEditingHabit(state);
  };

  return h<State>("div", {
    class: "py-8 rounded-4xl flex flex-col gap-8 overflow-hidden w-full",
    key: habit.id,
    style: { backgroundColor: HSLToString(color.background) }
  }, [
    h<State>("div", { class: "flex justify-between items-center px-8" }, [
      isEditing && isOwner
        ? h<State>("form", {
          class: "flex flex-col gap-4 w-full",
          onsubmit: handleEditSubmit,
          onkeydown: (state: State, event: KeyboardEvent) => {
            if (event.key === "Escape") {
              event.preventDefault();
              return CancelEditingHabit(state);
            }
            return state;
          },
        }, [
          h("input", {
            name: "habit-name",
            class: "text-xl font-bold py-2 bg-transparent focus:outline-none w-full",
            style: { borderBottom: `1px solid ${HSLToString(color.cellBorder)}` },
            value: editingHabit ? editingHabit.name : habit.name,
            autofocus: true,
            oninput: (state: State, event: Event) =>
              UpdateEditingHabitField(state, { field: 'name', value: (event.target as HTMLInputElement).value })
          }),
          h<State>("div", { class: "flex items-center justify-between" }, [
            h<State>("div", { class: "flex items-center gap-2" }, [
              h<State>("label", { for: `color-hue-${habit.id}`, class: "text-sm", style: { color: HSLToString(color.text) } }, text("Hue (0-360):")),
              h("input", {
                name: "habit-color",
                id: `color-hue-${habit.id}`,
                type: "number",
                min: "0",
                max: "360",
                value: editingHabit ? editingHabit.colorHue : habit.colorHue,
                class: "bg-transparent w-20 text-center py-2 focus:outline-none",
                style: { borderBottom: `1px solid ${HSLToString(color.cellBorder)}` },
                oninput: (state: State, event: Event) => {
                  const rawValue = (event.target as HTMLInputElement).value;
                  const value = parseInt(rawValue, 10);
                  if (!isNaN(value) && value >= 0 && value <= 360) {
                    return UpdateEditingHabitField(state, { field: 'colorHue', value });
                  }
                  return state;
                }
              }),
            ]),
            Button({ type: "submit", class: "py-1 px-3 text-sm" }, text("Save")),
          ])
        ])
        : h<State>("h3", { class: "text-xl font-bold" }, text(habit.name)),
      isOwner && !isEditing && token
        ? h("div", { class: "flex items-center gap-2" }, [
          habit.isBoolean
            ? h("button", {
              class: `p-2 rounded-full transition-colors ${wasLoggedToday
                ? 'bg-white/20 text-white hover:bg-white/40 outline-2 outline-white/40 outline-offset-2'
                : 'bg-white/10 hover:bg-white/30 text-white/30 hover:text-white'
                }`,
              onclick: (state: State) => [
                state,
                [UpsertHabitLogFx, { habitId: habit.id, date: todayStr, value: wasLoggedToday ? 0 : 1, token }]
              ],
              title: wasLoggedToday ? "Mark as not done for today" : "Mark as done for today"
            }, h("svg", { class: "w-6 h-6", fill: "none", viewBox: "0 0 24 24", "stroke-width": "2", stroke: "currentColor" },
              h("path", { "stroke-linecap": "round", "stroke-linejoin": "round", d: "M5 13l4 4L19 7" })
            ))
            : h("input", {
              type: "number",
              class: twMerge(
                "w-12 p-2 rounded-lg text-center bg-transparent focus:outline-none focus:ring-0 border-0 transition-colors no-spinners",
                wasLoggedToday
                  ? 'bg-white/20 text-white hover:bg-white/40 outline-2 outline-white/40 outline-offset-2'
                  : 'bg-white/10 hover:bg-white/30 text-white/30 hover:text-white'
              ),
              min: "0",
              value: todayValue,
              onchange: (state: State, event: Event) => {
                const target = event.target as HTMLInputElement;
                const rawValue = target.value;
                const value = rawValue === "" ? 0 : parseInt(rawValue, 10);

                if (!isNaN(value) && value >= 0) {
                  if (value !== todayValue && token) {
                    return [state, [UpsertHabitLogFx, { habitId: habit.id, date: todayStr, value, token }]];
                  }
                } else {
                  target.value = String(todayValue);
                }
                return state;
              },
              title: "Log value for today"
            }),

          h("div", { class: "relative" }, [
            h("button", {
              class: "p-2 rounded-full hover:bg-white/20 text-white hover:text-white transition-colors",
              onclick: [ToggleHabitMenu, habit.id]
            }, h("svg", { class: "w-6 h-6", xmlns: "http://www.w3.org/2000/svg", fill: "none", viewBox: "0 0 24 24", "stroke-width": "1.5", stroke: "currentColor" },
              h("path", { "stroke-linecap": "round", "stroke-linejoin": "round", d: "M12 6.75a.75.75 0 110-1.5.75.75 0 010 1.5zM12 12.75a.75.75 0 110-1.5.75.75 0 010 1.5zM12 18.75a.75.75 0 110-1.5.75.75 0 010 1.5z" })
            )),
            Popup({
              isOpen: activeHabitMenuId === habit.id,
              onClose: CloseHabitMenu,
              class: "top-full right-0 mt-2 w-48"
            }, HabitMenu({ habit, token }))
          ])
        ])
        : null
    ]),

    h<State>("div", { class: "relative" }, [
      h<State>("div", { class: "flex flex-nowrap flex-row-reverse w-full overflow-x-scroll overflow-y-hidden no-scrollbar items-end gap-4" },
        years.map(year =>
          h<State>("div", { key: String(year), class: "flex flex-col items-center gap-2 flex-shrink-0 px-8" }, [
            h<State>("div", { class: "font-bold", style: { color: HSLToString(color.text) } }, text(year)),
            CalendarGrid({ year, habit, logs: logsByYear[year] || [], isEditing, token, color }),
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
