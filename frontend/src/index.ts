import { h, text, app } from "hyperapp";
import type { State } from "./types";
import { initialState, initFx } from "./state";
import { AuthForm } from "./views/AuthForm";
import { DashboardView } from "./views/DashboardView";
import "./index.css";

const view = (state: State) => {
  switch (state.view) {
    case "login": return AuthForm({ isLogin: true })(state);
    case "signup": return AuthForm({ isLogin: false })(state);
    case "dashboard": return DashboardView(state);
    default: return h<State>("div", {}, text("Page not found"));
  }
};

// Initialize the app
app<State>({
  init: [initialState, [initFx, initialState]],
  view: view,
  node: document.getElementById("app")!,
});
