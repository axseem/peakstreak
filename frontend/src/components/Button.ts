import { h, type VNode } from "hyperapp"
import { twMerge } from "tailwind-merge";

type ButtonProps = {
  class?: string;
  [key: string]: unknown;
};

export const Button = <State>(
  props: ButtonProps,
  children: VNode<State> | VNode<State>[]
): VNode<State> => {
  const { class: customClass, ...rest } = props;
  const baseClass = "bg-white text-black text-nowrap hover:cursor-pointer rounded-lg py-2 px-4 disabled:cursor-not-allowed disabled:bg-neutral-600 disabled:text-neutral-900 enabled:hover:bg-neutral-100";

  return h(
    "button",
    {
      class: twMerge(baseClass, customClass),
      ...rest
    },
    children
  );
};
