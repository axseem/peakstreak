import { h, text, type VNode } from "hyperapp";
import type { State, FollowerListState } from "../types";
import { CloseFollowerList } from "../state";
import { UserListItem } from "./UserListItem";

export const UserListPopup = ({ followerList }: { followerList: FollowerListState }): VNode<State> | null => {
  if (!followerList.isOpen) {
    return null;
  }

  return h("div", {}, [
    h("div", {
      class: "fixed inset-0 bg-black/60 z-40",
      onclick: CloseFollowerList
    }),
    h("div", {
      class: "fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-50 bg-neutral-900 border border-neutral-800 rounded-xl shadow-lg w-full max-w-sm flex flex-col",
      onclick: (state: State, event: Event) => {
        event.stopPropagation();
        return state;
      }
    }, [
      h("header", { class: "flex items-center justify-between p-4 border-b border-neutral-800" }, [
        h("h2", { class: "text-xl font-bold" }, text(followerList.title)),
        h("button", {
          class: "p-1 rounded-full hover:bg-neutral-800",
          onclick: CloseFollowerList,
        },
          h("svg", { class: "w-6 h-6", xmlns: "http://www.w3.org/2000/svg", fill: "none", viewBox: "0 0 24 24", "stroke-width": "1.5", stroke: "currentColor" },
            h("path", { "stroke-linecap": "round", "stroke-linejoin": "round", d: "M6 18L18 6M6 6l12 12" })
          )
        )
      ]),

      h("div", { class: "p-4 max-h-[60vh] overflow-y-auto no-scrollbar" }, [
        followerList.isLoading
          ? h("p", { class: "text-center text-neutral-400" }, text("Loading..."))
          : followerList.error
            ? h("p", { class: "text-red-400" }, text(followerList.error))
            : followerList.users.length === 0
              ? h("p", { class: "text-center text-neutral-400" }, text("No users to show."))
              : h("div", { class: "flex flex-col gap-2" },
                followerList.users.map(user => UserListItem(user, CloseFollowerList))
              )
      ])
    ])
  ]);
};
