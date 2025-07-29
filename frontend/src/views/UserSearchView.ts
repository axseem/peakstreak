import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";
import { UserListItem } from "../components/UserListItem";
import { NavigateFx } from "../router";

export const UserSearchView = (state: State): VNode<State> => {
  const { query, results, isLoading, error } = state.search;

  return h("div", { class: "flex flex-col gap-6 w-full" }, [
    h(
      "form",
      {
        class: "relative",
        onsubmit: (s: State, e: Event) => {
          e.preventDefault();
          const formData = new FormData(e.target as HTMLFormElement);
          const newQuery = (formData.get("q") as string).trim();
          if (newQuery) {
            return [
              s,
              [
                NavigateFx,
                { path: `/search?q=${encodeURIComponent(newQuery)}` },
              ],
            ];
          }
          return s;
        },
      },
      [
        h(
          "span",
          { class: "absolute inset-y-0 left-0 flex items-center pl-4" },
          h("iconify-icon", {
            icon: "material-symbols:search",
            class: "text-neutral-500",
            width: 20,
            height: 20,
          }),
        ),
        h("input", {
          name: "q",
          placeholder: "Search users...",
          class:
            "w-full bg-neutral-900 border border-neutral-800 rounded-lg py-2.5 pl-11 pr-4 focus:outline-none placeholder-neutral-500",
          value: query,
          autofocus: true,
        }),
      ],
    ),

    query
      ? h("div", { class: "flex flex-col gap-4" }, [
          h("h1", { class: "text-2xl font-bold" }, [
            text("Search results for "),
            h("span", { class: "text-neutral-400" }, text(`"${query}"`)),
          ]),

          isLoading
            ? h("p", { class: "text-neutral-400" }, text("Searching..."))
            : error
              ? h("p", { class: "text-red-400" }, text(error))
              : results.length === 0
                ? h("p", { class: "text-neutral-400" }, text("No users found."))
                : h(
                    "div",
                    { class: "flex flex-col gap-2" },
                    results.map((user) => UserListItem(user)),
                  ),
        ])
      : h(
          "p",
          { class: "text-center text-neutral-500 mt-4" },
          text("Find new people to follow."),
        ),
  ]);
};
