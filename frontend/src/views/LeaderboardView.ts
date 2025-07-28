import { h, text, type VNode } from "hyperapp";
import type { State, LeaderboardEntry } from "../types";
import { HabitCard } from "../components/HabitCard";
import { Avatar } from "../components/Avatar";
import { NavigateFx } from "../router";

const LeaderboardUserCard = ({ entry, rank }: { entry: LeaderboardEntry, rank: number }): VNode<State> => {
  return h("div", {
    class: "flex flex-col gap-4 p-6 rounded-2xl bg-neutral-900/50 border border-neutral-800/50"
  }, [
    h("div", { class: "flex items-center justify-between gap-4" }, [
      h("div", { class: "flex items-center gap-4 flex-grow min-w-0" }, [
        h("span", { class: "text-2xl font-bold text-neutral-500 w-8 text-center flex-shrink-0" }, text(`${rank}`)),
        h("a", {
          href: `/@${entry.user.username}`,
          class: "flex items-center gap-3 min-w-0",
          onclick: (s: State, e: Event) => {
            e.preventDefault();
            return [s, [NavigateFx, { path: `/@${entry.user.username}` }]];
          }
        }, [
          Avatar({ src: entry.user.avatarUrl, username: entry.user.username, sizeClass: "w-12 h-12" }),
          h("span", { class: "font-bold text-lg truncate" }, text(entry.user.username)),
        ]),
      ]),
      h("div", { class: "flex flex-col items-end flex-shrink-0" }, [
        h("span", { class: "text-xl font-bold" }, text(entry.totalLoggedDays)),
        h("span", { class: "text-sm text-neutral-400" }, text("days logged")),
      ]),
    ]),
    entry.habits.length > 0 && h("div", { class: "flex flex-col gap-8 pt-4 mt-4 border-t border-neutral-800/50" },
      entry.habits.map(habit => HabitCard({ habit, isOwner: false, token: null, isEditing: false, activeHabitMenuId: null }))
    )
  ]);
};

export const LeaderboardView = (state: State): VNode<State> => {
  const { users, isLoading, error } = state.leaderboard;

  if (isLoading) {
    return h("p", {}, text("Loading leaderboard..."));
  }

  if (error) {
    return h("p", { class: "text-red-400" }, text(error));
  }

  return h("div", { class: "flex flex-col gap-8 w-full" }, [
    h("h1", { class: "text-3xl font-bold" }, text("Leaderboard")),
    users.length === 0
      ? h("p", { class: "text-neutral-500" }, text("The leaderboard is empty. Start tracking habits!"))
      : h("div", { class: "flex flex-col gap-6" },
        users.map((entry, index) => LeaderboardUserCard({ entry, rank: index + 1 }))
      ),
  ]);
};
