import { h, text, type VNode } from "hyperapp"

type NativeInputProps = Omit<Record<string, unknown>, "id"> & { id: string };
type InputProps = NativeInputProps & {
  label?: string;
};

const nativeInput = <State>(props: NativeInputProps): VNode<State> => h("input", {
  class: "bg-neutral-800 rounded p-2",
  ...props,
});

export const Input = <State>({ label, id, ...rest }: InputProps): VNode<State> => {
  if (label === undefined) {
    return nativeInput({ id, ...rest });
  }

  return h("div", { class: "flex flex-col gap-2" },
    [
      h("label", { for: id }, text(label)),
      nativeInput({ id, name: id, ...rest }),
    ]
  )
}
