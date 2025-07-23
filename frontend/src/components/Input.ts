import { h, text, type VNode } from "hyperapp"
import { twMerge } from "tailwind-merge";
import type { State } from "../types";

type NativeInputProps = Omit<Record<string, unknown>, "id"> & {
  id: string;
  class?: string;
};
type InputProps = NativeInputProps & {
  label?: string;
};

const nativeInput = ({ class: customClass, ...rest }: NativeInputProps): VNode<State> => {
  const baseClass = "bg-neutral-800 rounded-lg py-2 px-4 focus:outline-none min-w-0 flex-grow";

  return h("input", {
    class: twMerge(baseClass, customClass),
    ...rest,
  });
};

export const Input = ({ label, id, ...rest }: InputProps): VNode<State> => {
  if (label === undefined) {
    return nativeInput({ id, ...rest });
  }

  return h("div", { class: "flex flex-col gap-2 min-w-0 w-full" },
    [
      h("label", { for: id }, text(label)),
      nativeInput({ id, name: id, ...rest }),
    ]
  )
}
