import { h, type VNode } from "hyperapp";
import type { State } from "../types";
import { Sidebar } from "../components/Sidebar";

export const MainLayout = (state: State, child: VNode<State>): VNode<State> => {
  return h("div", { class: "flex justify-center min-h-screen bg-black" }, [
    Sidebar(state),
    h("main", { class: "w-full max-w-2xl p-4 pb-20 md:p-8" }, child),
  ]);
};
