import { h, text, type VNode } from "hyperapp"

type NativeInputProps = Omit<Record<string, unknown>, "id"> & {
  id: string;
  class?: string;
};
type InputProps = NativeInputProps & {
  label?: string;
};

const nativeInput = <State>({ class: additionalClass, ...rest }: NativeInputProps): VNode<State> => {
  const baseClass = "bg-neutral-800 rounded-lg py-2 px-4 focus:outline-none min-w-0 flex-grow";
  const finalClass = [baseClass, additionalClass].filter(Boolean).join(" ");

  return h("input", {
    class: finalClass,
    ...rest,
  });
};

export const Input = <State>({ label, id, ...rest }: InputProps): VNode<State> => {
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
