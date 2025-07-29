import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";

type AvatarProps = {
  src?: string | null;
  username?: string;
  sizeClass: string;
  class?: string;
};

export const Avatar = ({
  src,
  username,
  sizeClass,
  class: customClass,
}: AvatarProps): VNode<State> => {
  const baseClass = "rounded-full object-cover flex-shrink-0";
  const containerClass = `${baseClass} ${sizeClass} ${customClass || ""}`;

  if (src) {
    return h("img", {
      src,
      class: containerClass,
      alt: username ? `${username}'s avatar` : "avatar",
    });
  }

  const initial = username ? username.charAt(0).toUpperCase() : "?";
  const placeholderClass = `bg-neutral-700 flex items-center justify-center text-white/50 font-bold overflow-hidden select-none leading-none ${containerClass}`;
  const fontSize = `calc(${sizeClass.split(" ")[0].replace("w-", "")}rem / 6)`;

  return h(
    "div",
    { class: placeholderClass, title: username },
    h("span", { style: { fontSize } }, text(initial)),
  );
};
