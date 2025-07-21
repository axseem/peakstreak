import { h, text, app } from "hyperapp";
import type { State } from "./types";
import { initialState, initFx } from "./state";
import { AuthForm } from "./views/AuthForm";
import { DashboardView } from "./views/DashboardView";
import { MainLayout } from "./views/MainLayout";
import { PopStateSub } from "./router";
import "./index.css";

const ComingSoon = (page: string) => h<State>("div", { class: "w-full h-full flex items-center justify-center text-2xl text-neutral-500" }, text(`${page} - Coming Soon!`));

const renderView = (state: State) => {
  switch (state.view) {
    case "login": return AuthForm({ isLogin: true })(state);
    case "signup": return AuthForm({ isLogin: false })(state);
    case "dashboard": return DashboardView(state);
    case "leaderboard": return ComingSoon("Leaderboard");
    case "explore": return ComingSoon("Explore");
    case "create_habit": return ComingSoon("Create Habit");
    case "not_found":
    default:
      return h<State>("div", { class: "w-full h-screen flex items-center justify-center" }, text("404 - Page Not Found"));
  }
}

const view = (state: State) => {
  const viewsWithoutSidebar: State["view"][] = ["login", "signup"];
  const page = renderView(state);

  if (viewsWithoutSidebar.includes(state.view) || !state.token) {
    return page;
  }

  return MainLayout(state, page);
}

app<State>({
  init: [initialState, [initFx, initialState]],
  view: view,
  subscriptions: (state) => [
    state && [PopStateSub, {}]
  ],
  node: document.getElementById("app")!,
});
