import { h, text, type VNode } from "hyperapp";
import type { State, PublicUser } from "../types";
import { NavigateFx } from "../router";
import { Avatar } from "./Avatar";

export const UserListItem = (
  user: PublicUser,
  onclickAction?: (state: State) => State,
): VNode<State> =>
  h(
    "a",
    {
      href: `/@${user.username}`,
      class:
        "flex items-center gap-4 p-4 rounded-lg hover:bg-neutral-800 transition-colors w-full",
      onclick: (state: State, event: Event) => {
        event.preventDefault();
        const nextState = onclickAction ? onclickAction(state) : state;
        return [nextState, [NavigateFx, { path: `/@${user.username}` }]];
      },
    },
    [
      Avatar({
        src: user.avatarUrl,
        username: user.username,
        sizeClass: "w-12 h-12",
      }),
      h("div", { class: "flex flex-col" }, [
        h("span", { class: "font-bold text-lg" }, text(user.username)),
      ]),
    ],
  );
