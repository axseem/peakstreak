import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";
import { Logout, HandleFormInput, CreateHabitFx, LogHabitFx } from "../state";

export const DashboardView = (state: State): VNode<State> =>
  h<State>("div", { class: "dashboard-container" }, [
    h<State>("header", { class: "header" }, [
      h<State>("h1", {}, text(`Welcome, ${state.user?.username}`)),
      h<State>("button", { class: "button-logout", onclick: Logout }, text("Logout")),
    ]),
    h<State>("div", { class: "form-container" }, [
      h<State>("h2", { class: "w-80 bg-black text-white" }, text("Create a New Habit")),
      h<State>("form", {
        onsubmit: (state: State, event: Event) => {
          event.preventDefault();
          const { newHabitName, token, user } = state;
          if (newHabitName.trim() && token && user) {
            return [
              { ...state, newHabitName: '' },
              [CreateHabitFx, { name: newHabitName, token, username: user.username }]
            ];
          }
          return state;
        }
      }, [
        h<State>("input", {
          placeholder: "e.g., Read for 15 minutes",
          value: state.newHabitName,
          oninput: HandleFormInput,
          disabled: state.isLoading,
        }),
        h<State>("button", { type: "submit", disabled: state.isLoading || !state.newHabitName.trim() }, text("Add Habit")),
      ])
    ]),
    h<State>("div", { class: "habit-list-container" }, [
      h<State>("h2", {}, text("Your Habits")),
      state.isLoading && state.habits.length === 0
        ? h<State>("p", {}, text("Loading habits..."))
        : h<State>("div", { class: "habit-list" }, state.habits.map(habit =>
          h<State>("div", { class: "habit-item", key: habit.id }, [
            h<State>("span", { class: "habit-item-name" }, text(habit.name)),
            h<State>("button", {
              class: `habit-log-button ${habit.loggedToday ? 'logged' : ''}`,
              disabled: state.isLoading || habit.loggedToday,
              onclick: (state: State) => [state, [LogHabitFx, { habitId: habit.id, token: state.token! }]]
            }, text(habit.loggedToday ? "Logged!" : "Log for Today"))
          ])
        ))
    ])
  ]);
