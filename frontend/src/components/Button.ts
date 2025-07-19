import { h, text as t, type VNode } from "hyperapp"

type InputProps = Record<string, unknown> & {
  text: string;
};

export const Button = <State>({ text, ...rest }: InputProps): VNode<State> => {
  return h("button", {
    class: "bg-white text-black hover:cursor-pointer rounded p-2",
    ...rest
  }, t(text))
};
