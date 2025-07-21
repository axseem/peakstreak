// Path: frontend/src/views/ProfileView.ts
// (This file replaces DashboardView.ts)
import { h, text, type VNode } from "hyperapp";
import type { State, HabitWithLogs, HabitLog } from "../types";
import { HandleFormInput, CreateHabitFx, LogHabitFx } from "../state";
import { toYYYYMMDD, getDatesForYear, groupLogsByYear } from "../lib/date";
import { Input } from "../components/Input";
import { Button } from "../components/Button";

const YearTable = ({ year, logs }: { year: number, logs: HabitLog[] }): VNode<State> | null => {
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
    style: { borderSpacing: '6px' }
  }, [
    h("tbody", {},
      calendarGrid.map(weekRow =>
        h("tr", {},
          weekRow.slice(0, maxWeekOfYear + 1).map((day, i) => {
            if (!day) {
              return h("td", { key: `empty-${i}`, class: "w-3 h-3" });
            }
            const isLogged = logsMap.get(toYYYYMMDD(day)) === true;
            return h("td", {
              key: toYYYYMMDD(day),
              class: `w-3 h-3 rounded-sm ${isLogged ? 'bg-teal-500' : 'bg-neutral-800'}`,
              title: `${day.toDateString()}: ${isLogged ? 'Completed' : 'Not completed'}`
            });
          })
        )
      )
    )
  ]);
};

const HabitCard = (habit: HabitWithLogs, isOwner: boolean, token: string | null): VNode<State> => {
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

  return h("div", { class: "bg-neutral-900 py-8 rounded-4xl flex flex-col gap-8 overflow-hidden w-full", key: habit.id }, [
    h("div", { class: "flex justify-between items-center px-8" }, [
      h("h3", { class: "text-xl font-bold" }, text(habit.name)),
      isOwner && token
        ? Button({
          disabled: wasLoggedToday,
          onclick: (state: State) => [state, [LogHabitFx, { habitId: habit.id, token: state.token! }]],
          class: "text-xs"
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
            YearTable({ year, logs: logsByYear[year] || [] }),
          ])
        )
      ),
      h("div", { class: "absolute inset-y-0 left-0 w-8 bg-gradient-to-r from-neutral-900 to-transparent pointer-events-none" }),
      h("div", { class: "absolute inset-y-0 right-0 w-8 bg-gradient-to-l from-neutral-900 to-transparent pointer-events-none" })
    ])
  ]);
};

const CreateHabitForm = (state: State): VNode<State> => h("div", {}, [
  h("form", {
    class: "flex items-end gap-2",
    onsubmit: (state: State, event: Event) => {
      event.preventDefault();
      if (state.newHabitName.trim() && state.token) {
        return [
          state,
          [CreateHabitFx, { name: state.newHabitName, token: state.token }]
        ];
      }
      return state;
    }
  }, [
    Input({
      label: "Create new habit",
      id: "newHabit",
      placeholder: "Read for 15 minutes",
      value: state.newHabitName,
      oninput: HandleFormInput,
      disabled: state.isLoading,
    }),
    Button({
      type: "submit",
      disabled: state.isLoading || !state.newHabitName.trim(),
    }, text("Add Habit")),
  ])
]);

export const ProfileView = (state: State): VNode<State> => {
  if (state.isLoading && !state.profileData) {
    return h("p", {}, text("Loading profile..."));
  }

  if (!state.profileData) {
    return h("p", { class: "text-red-400" }, text(state.error || "Profile could not be loaded."));
  }

  const { user, habits, isOwner } = state.profileData;

  return h("div", { class: "flex flex-col gap-12 w-full" }, [
    h("header", { class: "flex justify-between items-center" }, [
      h("h1", { class: "text-3xl font-bold" }, text(user.username)),
    ]),

    isOwner ? CreateHabitForm(state) : null,

    h("div", { class: "flex flex-col gap-8" },
      habits.length > 0
        ? habits.map(habit => HabitCard(habit, isOwner, state.token))
        : h("p", { class: "text-neutral-400" }, text(isOwner ? "You don't have any habits yet. Add one above to get started!" : "This user hasn't added any habits yet."))
    )
  ]);
};
