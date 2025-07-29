import { h, type VNode, type Dispatchable } from "hyperapp";
import type { State } from "../types";
import { twMerge } from "tailwind-merge";

type PopupProps = {
  isOpen: boolean;
  onClose: (state: State) => Dispatchable<State, any>;
  class?: string;
};

export const Popup = (
  props: PopupProps,
  children: VNode<State> | VNode<State>[],
): VNode<State> | null => {
  if (!props.isOpen) {
    return null;
  }

  const contentBaseClass =
    "bg-neutral-900 border border-neutral-800 rounded-lg shadow-lg";

  return h("div", {}, [
    h("div", {
      class: "fixed inset-0 z-40",
      onclick: props.onClose,
    }),

    h(
      "div",
      {
        class: twMerge("absolute z-50", contentBaseClass, props.class),
        onclick: (state: State, event: Event) => {
          event.stopPropagation();
          return state;
        },
      },
      children,
    ),
  ]);
};
