import { h, text, type VNode } from "hyperapp";
import type { State, ExploreEntry } from "../types";
import { HabitCard } from "../components/HabitCard";
import { Avatar } from "../components/Avatar";
import { NavigateFx } from "../router";

const ExploreUserCard = (entry: ExploreEntry): VNode<State> => {
  return h("div", {
    class: "flex flex-col gap-4 p-6 rounded-2xl bg-neutral-900/50 border border-neutral-800/50"
  }, [
    h("div", { class: "flex items-center gap-4" }, [
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
    h("div", { class: "pt-4 mt-4 border-t border-neutral-800/50" },
      HabitCard({ habit: entry.habit, isOwner: false, token: null, isEditing: false, activeHabitMenuId: null, editingHabit: null })
    )
  ]);
};

export const ExploreView = (state: State): VNode<State> => {
  const { entries, isLoading, error } = state.explore;

  if (isLoading) {
    return h("p", {}, text("Loading explore page..."));
  }

  if (error) {
    return h("p", { class: "text-red-400" }, text(error));
  }

  return h("div", { class: "flex flex-col gap-8 w-full" }, [
    h("h1", { class: "text-3xl font-bold" }, text("Explore")),
    entries.length === 0
      ? h("p", { class: "text-neutral-500" }, text("Nothing to see here yet. Start logging habits!"))
      : h("div", { class: "flex flex-col gap-6" },
        entries.map((entry) => ExploreUserCard(entry))
      ),
  ]);
};
