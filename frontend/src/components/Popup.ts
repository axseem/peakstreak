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
  children: VNode<State> | VNode<State>[]
): VNode<State> | null => {
  if (!props.isOpen) {
    return null;
  }

  const contentBaseClass = "bg-neutral-900 border border-neutral-800 rounded-lg shadow-lg";

  // The component returns a single, unstyled wrapper `div` so we can have two sibling elements.
  return h("div", {}, [
    // 1. The full-screen overlay for "click outside to close".
    // It's fixed to the viewport and has a lower z-index.
    h("div", {
      class: "fixed inset-0 z-40",
      onclick: props.onClose,
    }),

    // 2. The actual popup content. It's a sibling to the overlay.
    // It will be positioned absolutely relative to its parent container.
    h("div", {
      class: twMerge("absolute z-50", contentBaseClass, props.class),
      onclick: (state: State, event: Event) => {
        // Stop clicks inside the popup from closing it.
        event.stopPropagation();
        return state;
      },
    }, children),
  ]);
};
