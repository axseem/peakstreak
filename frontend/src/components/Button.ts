import { h, type VNode } from "hyperapp"

type ButtonProps = Record<string, unknown>;

export const Button = <State>(
  props: ButtonProps,
  children: VNode<State> | VNode<State>[]
): VNode<State> => {
  const { class: customClass, ...rest } = props;
  const finalClass = [
    "bg-white text-black text-nowrap hover:cursor-pointer rounded-lg py-2 px-4 disabled:cursor-not-allowed disabled:bg-neutral-600 disabled:text-neutral-900 enabled:hover:bg-neutral-100",
    customClass
  ].filter(Boolean).join(" ");

  return h(
    "button",
    {
      class: finalClass,
      ...rest
    },
    children
  );
};
