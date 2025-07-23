import { h, text, type VNode, type Action } from "hyperapp";
import type { State, HabitWithLogs, HabitLog } from "../types";
import {
  HandleFormInput,
  CreateHabitFx,
  LogHabitFx,
  ShowAddHabitForm,
  HideAddHabitForm,
  ToggleEditMode,
  UpdateHabitNameFx,
  ToggleHabitLogFx,
  FollowUserFx,
  UnfollowUserFx
} from "../state";
import { toYYYYMMDD, getDatesForYear, groupLogsByYear } from "../lib/date";
import { Input } from "../components/Input";
import { Button } from "../components/Button";

const YearTable = ({ year, logs, isEditing, habitId, token }: { year: number, logs: HabitLog[], isEditing: boolean, habitId: string, token: string | null }): VNode<State> | null => {
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
            const dateStr = toYYYYMMDD(day);
            const isLogged = logsMap.get(dateStr) === true;
            const canToggle = isEditing && new Date(new Date().setUTCHours(0, 0, 0, 0)) >= day;

            return h("td", {
              key: dateStr,
              class: `w-3 h-3 rounded-sm ${isLogged ? 'bg-teal-500' : 'bg-neutral-800'} ${canToggle ? 'cursor-pointer hover:opacity-75' : ''}`,
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

const HabitCard = (habit: HabitWithLogs, isOwner: boolean, token: string | null, isEditing: boolean): VNode<State> => {
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
      return [state, [UpdateHabitNameFx, { habitId: habit.id, name: newName, token }]];
    }
    target.value = habit.name; // Revert if invalid or unchanged
    return state;
  };

  return h("div", { class: "bg-neutral-900 py-8 rounded-4xl flex flex-col gap-8 overflow-hidden w-full", key: habit.id }, [
    h("div", { class: "flex justify-between items-center px-8" }, [
      isEditing && isOwner
        ? h("input", {
          class: "text-xl font-bold bg-transparent border-b border-neutral-700 focus:outline-none w-full mr-4",
          value: habit.name,
          onblur: onNameChange,
          onkeydown: (state: State, event: KeyboardEvent) => {
            if (event.key === "Enter") (event.target as HTMLInputElement).blur();
            return state;
          },
        })
        : h("h3", { class: "text-xl font-bold" }, text(habit.name)),
      isOwner && !isEditing && token
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
            YearTable({ year, logs: logsByYear[year] || [], isEditing, habitId: habit.id, token }),
          ])
        )
      ),
      h("div", { class: "absolute inset-y-0 left-0 w-8 bg-gradient-to-r from-neutral-900 to-transparent pointer-events-none" }),
      h("div", { class: "absolute inset-y-0 right-0 w-8 bg-gradient-to-l from-neutral-900 to-transparent pointer-events-none" })
    ])
  ]);
};

const InlineCreateHabitForm = (state: State): VNode<State> => h("form", {
  class: "flex flex-col items-center justify-center gap-4 p-8 w-full h-full",
  onsubmit: (s: State, event: Event) => {
    event.preventDefault();
    if (s.newHabitName.trim() && s.token) {
      return [s, [CreateHabitFx, { name: s.newHabitName, token: s.token }]];
    }
    return s;
  },
}, [
  Input({
    label: "What habit do you want to build?",
    id: "newHabitInline",
    placeholder: "e.g. Read for 15 minutes",
    value: state.newHabitName,
    oninput: HandleFormInput,
    disabled: state.isLoading,
    oncreate: (el: HTMLInputElement) => el.focus(),
    class: "text-center"
  }),
  h("div", { class: "flex gap-2 w-full mt-4" }, [
    Button({
      type: "button",
      onclick: HideAddHabitForm,
      class: "w-full bg-neutral-700 enabled:hover:bg-neutral-800 text-neutral-300"
    }, text("Cancel")),
    Button({
      type: "submit",
      disabled: state.isLoading || !state.newHabitName.trim(),
      class: "w-full"
    }, text("Add Habit")),
  ])
]);

const AddHabitCard = (state: State): VNode<State> => {
  if (state.isAddingHabit) {
    return h("div", {
      class: "bg-neutral-900 rounded-4xl flex items-center justify-center w-full",
      key: "add-habit-form"
    }, [
      InlineCreateHabitForm(state)
    ]);
  }

  return h("div", {
    key: "add-habit-button",
    class: "bg-transparent border-2 border-dashed border-neutral-800 rounded-4xl flex items-center justify-center min-h-[16rem] hover:bg-neutral-900 hover:border-neutral-700 transition-colors cursor-pointer",
    onclick: ShowAddHabitForm
  }, [
    h("svg", {
      xmlns: "http://www.w3.org/2000/svg",
      fill: "none",
      viewBox: "0 0 24 24",
      "stroke-width": 1.5,
      stroke: "currentColor",
      class: "w-12 h-12 text-neutral-700"
    }, [
      h("path", {
        "stroke-linecap": "round",
        "stroke-linejoin": "round",
        d: "M12 4.5v15m7.5-7.5h-15"
      })
    ])
  ]);
};

export const ProfileView = (state: State): VNode<State> => {
  if (state.isLoading && !state.profileData) {
    return h("p", {}, text("Loading profile..."));
  }

  if (!state.profileData) {
    return h("p", { class: "text-red-400" }, text(state.error || "Profile could not be loaded."));
  }

  const { user, habits, isOwner, followersCount, followingCount, isFollowing } = state.profileData;

  const FollowButton = () => Button({
    onclick: (s: State) => {
      if (s.isLoading || !s.token || !s.profileData) return s;
      const fx = s.profileData.isFollowing ? UnfollowUserFx : FollowUserFx;
      return [s, [fx, { username: s.profileData.user.username, token: s.token }]];
    },
    class: isFollowing ? "bg-neutral-700 enabled:hover:bg-neutral-800 text-neutral-300" : "",
    disabled: state.isLoading,
  }, text(isFollowing ? "Following" : "Follow"));

  const EditButton = () => Button({
    onclick: ToggleEditMode
  }, text(state.isEditing ? "Done" : "Edit"));


  return h("div", { class: "flex flex-col gap-8 w-full" }, [
    h("header", { class: "flex justify-between items-start" }, [
      h("div", { class: "flex flex-col gap-2" }, [
        h("h1", { class: "text-3xl font-bold" }, text("@" + user.username)),
        h("div", { class: "flex gap-4 text-neutral-400" }, [
          h("span", {}, text(`${followersCount} Followers`)),
          h("span", {}, text(`${followingCount} Following`)),
        ]),
      ]),
      isOwner
        ? (habits.length > 0 && EditButton())
        : FollowButton()
    ]),

    h("div", { class: "flex flex-col gap-4" }, [
      h("div", { class: "flex justify-between items-center" }, [
        h("h2", { class: "text-2xl font-bold" }, text("Habits")),
      ]),
      h("div", { class: "flex flex-col gap-8" }, [
        ...habits.map(habit => HabitCard(habit, isOwner, state.token, state.isEditing)),

        isOwner && !state.isEditing ? AddHabitCard(state) : null,

        habits.length === 0 && !isOwner
          ? h("p", { class: "text-neutral-400" }, text("This user hasn't added any habits yet."))
          : null
      ])
    ])
  ]);
};
