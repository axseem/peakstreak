import { h, text, app } from "hyperapp";
import type { State } from "./types";
import { initialState, initFx } from "./state";
import { AuthForm } from "./views/AuthForm";
import { DashboardView } from "./views/DashboardView";
import { PopStateSub } from "./router";
import "./index.css";

const view = (state: State) => {
  switch (state.view) {
    case "login": return AuthForm({ isLogin: true })(state);
    case "signup": return AuthForm({ isLogin: false })(state);
    case "dashboard": return DashboardView(state);
    default:
      return h<State>("div", {}, text(`Page not found for view: ${state.view}`));
  }
};

app<State>({
  init: [initialState, [initFx, initialState]],
  view: view,
  subscriptions: (state) => [
    state && [PopStateSub, {}]
  ],
  node: document.getElementById("app")!,
});
