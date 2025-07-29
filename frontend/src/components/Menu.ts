import { h, type VNode, type Dispatchable } from "hyperapp";
import { twMerge } from "tailwind-merge";
import type { State } from "../types";

type MenuItemProps = {
  onclick: (state: State) => Dispatchable<State, any>;
  class?: string;
  confirmation?: string;
};

export const MenuItem = (props: MenuItemProps, children: any): VNode<State> => {
  return h(
    "button",
    {
      class: twMerge(
        "w-full text-left px-3 py-2 text-sm text-neutral-300 hover:bg-neutral-800 flex items-center gap-3 transition-colors rounded-md",
        props.class,
      ),
      onclick: (state: State) => {
        if (props.confirmation && !window.confirm(props.confirmation)) {
          return state;
        }
        return props.onclick(state);
      },
    },
    children,
  );
};

type MenuProps = {
  class?: string;
};

export const Menu = (props: MenuProps, children: any): VNode<State> => {
  return h(
    "div",
    {
      class: twMerge("p-1", props.class),
    },
    children,
  );
};
