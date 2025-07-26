import { h, text, type VNode, type Action } from "hyperapp";
import type { State, HabitWithLogs, HabitLog } from "../types";
import {
  LogHabitFx,
  UpdateHabitFx,
  ToggleHabitLogFx,
} from "../state";
import { toYYYYMMDD, getDatesForYear, groupLogsByYear } from "../lib/date";
import { Button } from "../components/Button";

type HSL = {
  hue: number,
  saturation: number,
  lightness: number,
}

type Color = {
  background: HSL,
  monthBackground: HSL,
  cellBorder: HSL,
  cell: HSL,
  text: HSL,
}

const HSLToString = (color: HSL): string => {
  return `hsl(${color.hue} ${color.saturation}% ${color.lightness}%)`;
}

const getColor = (hue: number): Color => {
  return {
    background: { hue, saturation: 54, lightness: 13 },
    monthBackground: { hue, saturation: 60, lightness: 19 },
    cellBorder: { hue, saturation: 54, lightness: 29 },
    cell: { hue, saturation: 100, lightness: 60 },
    text: { hue, saturation: 60, lightness: 50 },
  };
}

const YearTable = ({ year, logs, isEditing, habitId, token, color }: { year: number, logs: HabitLog[], isEditing: boolean, habitId: string, token: string | null, color: Color }): VNode<State> | null => {
  const logsMap = new Map(logs.map(log => [toYYYYMMDD(new Date(log.date)), log.status]));
  const dates = getDatesForYear(year);

  if (dates.length === 0) {
    return null;
  }

  const calendarGrid: (Date | null)[][] = Array.from({ length: 7 }, () => Array(53).fill(null));

  const yearStart = new Date(Date.UTC(year, 0, 1));
  const firstDayOfWeek = yearStart.getUTCDay();
  let maxWeekOfYear = 0;

  dates.forEach(date => {
    const dayOfYear = Math.floor((date.getTime() - yearStart.getTime()) / 86400000);
    const dayOfWeek = date.getUTCDay();
    const weekOfYear = Math.floor((dayOfYear + firstDayOfWeek) / 7);

    if (weekOfYear < 53) {
      calendarGrid[dayOfWeek][weekOfYear] = date;
      if (weekOfYear > maxWeekOfYear) {
        maxWeekOfYear = weekOfYear;
      }
    }
  });

  return h("table", {
    class: "border-separate table-fixed",
    style: { borderSpacing: '5px' }
  }, [
    h("tbody", {},
      calendarGrid.map(weekRow =>
        h("tr", {},
          weekRow.slice(0, maxWeekOfYear + 1).map((day, i) => {
            if (!day) {
              return h("td", { key: `empty-${i}`, class: "w-3 h-3" });
            }
            const dateStr = toYYYYMMDD(day);
            const isLogged = logsMap.get(dateStr) === true;
            const canToggle = isEditing && new Date(new Date().setUTCHours(0, 0, 0, 0)) >= day;

            return h("td", {
              key: dateStr,
              class: `w-3 h-3 rounded-sm ${canToggle ? 'cursor-pointer hover:opacity-75' : ''}`,
              style: {
                backgroundColor: isLogged
                  ? HSLToString(color.cell)
                  : "",
                border: isLogged
                  ? ""
                  : `1px solid, ${HSLToString(color.cellBorder)}`,
              },
              title: `${day.toDateString()}: ${isLogged ? 'Completed' : 'Not completed'}`,
              onclick: (state: State) => {
                if (!canToggle || !token) return state;
                return [state, [ToggleHabitLogFx, { habitId, date: dateStr, currentStatus: isLogged, token }]];
              }
            });
          })
        )
      )
    )
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
    target.value = habit.name; // Revert if invalid or unchanged
    return state;
  };

  const onColorHueChange: Action<State, Event> = (state, event) => {
    const target = event.target as HTMLInputElement;
    const newColorHue = parseInt(target.value, 10);
    if (!isNaN(newColorHue) && newColorHue >= 0 && newColorHue <= 360 && newColorHue !== habit.colorHue && token) {
      return [state, [UpdateHabitFx, { habitId: habit.id, name: habit.name, colorHue: newColorHue, token }]];
    }
    // Revert if invalid
    target.value = String(habit.colorHue);
    return state;
  };

  return h("div", {
    class: "py-8 rounded-4xl flex flex-col gap-8 overflow-hidden w-full",
    key: habit.id,
    style: {
      backgroundColor: HSLToString(color.background),
    }
  }, [
    h("div", { class: "flex justify-between items-center px-8" }, [
      isEditing && isOwner
        ? h("div", { class: "flex flex-col gap-2 w-full" }, [
          h("input", {
            class: "text-xl font-bold py-2 bg-transparent focus:outline-none w-full",
            style: {
              borderBottom: `1px solid ${HSLToString(color.cellBorder)}`
            },
            value: habit.name,
            onblur: onNameChange,
            onkeydown: (state: State, event: KeyboardEvent) => {
              if (event.key === "Enter") (event.target as HTMLInputElement).blur();
              return state;
            },
          }),
          h("div", { class: "flex items-center gap-2" }, [
            h("label", {
              for: `color-hue-${habit.id}`,
              class: "text-sm",
              style: {
                color: HSLToString(color.text)
              },
            }, text("Hue (0-360):")),
            h("input", {
              id: `color-hue-${habit.id}`,
              type: "number",
              min: "0",
              max: "360",
              value: habit.colorHue,
              class: "bg-transparent w-20 text-center py-2 focus:outline-none",
              style: {
                borderBottom: `1px solid ${HSLToString(color.cellBorder)}`
              },
              onblur: onColorHueChange,
              onkeydown: (state: State, event: KeyboardEvent) => {
                if (event.key === "Enter") (event.target as HTMLInputElement).blur();
                return state;
              },
            }),
          ])
        ])
        : h("h3", { class: "text-xl font-bold" }, text(habit.name)),
      isOwner && !isEditing && token
        ? Button({
          disabled: wasLoggedToday,
          onclick: (state: State) => [state, [LogHabitFx, { habitId: habit.id, token: state.token! }]],
          class: "text-xs",
          style: {
            backgroundColor: HSLToString(color.text)
          }
        }, text(wasLoggedToday ? "Completed!" : "Log Today"))
        : null
    ]),

    h("div", { class: "relative" }, [
      h("div", { class: "flex flex-nowrap flex-row-reverse w-full overflow-x-scroll no-scrollbar items-end gap-4" },
        years.map(year =>
          h("div", {
            key: String(year),
            class: "flex flex-col items-center gap-2 flex-shrink-0 px-8"
          }, [
            YearTable({ year, logs: logsByYear[year] || [], isEditing, habitId: habit.id, token, color }),
          ])
        )
      ),
      h("div", {
        class: "absolute inset-y-0 left-0 w-8 pointer-events-none",
        style: {
          background: `linear-gradient(to right, ${HSLToString(color.background)}, hsla(${color.background.hue} ${color.background.saturation}% ${color.background.lightness}% / 0))`,
        }
      }),
      h("div", {
        class: "absolute inset-y-0 right-0 w-8 pointer-events-none",
        style: {
          background: `linear-gradient(to left, ${HSLToString(color.background)}, hsla(${color.background.hue} ${color.background.saturation}% ${color.background.lightness}% / 0))`,
        }
      })
    ])
  ]);
};
