import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";
import { UserListItem } from "../components/UserListItem";

export const UserSearchView = (state: State): VNode<State> => {
  const { query, results, isLoading, error } = state.search;

  return h("div", { class: "flex flex-col gap-6 w-full" }, [
    h("h1", { class: "text-3xl font-bold" }, [
      text("Search results for "),
      h("span", { class: "text-neutral-400" }, text(`"${query}"`))
    ]),

    isLoading
      ? h("p", { class: "text-neutral-400" }, text("Searching..."))
      : error
        ? h("p", { class: "text-red-400" }, text(error))
        : results.length === 0
          ? h("p", { class: "text-neutral-400" }, text("No users found."))
          : h("div", { class: "flex flex-col gap-2" },
            results.map(user => UserListItem(user))
          )
  ]);
};
