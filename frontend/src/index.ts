import { h, text, app } from "hyperapp";
import type { State } from "./types";
import { initialState, initFx } from "./state";
import { AuthForm } from "./views/AuthForm";
import { ProfileView } from "./views/ProfileView";
import { MainLayout } from "./views/MainLayout";
import { PopStateSub } from "./router";
import "./index.css";
import { UserSearchView } from "./views/UserSearchView";
import { LeaderboardView } from "./views/LeaderboardView";
import { ExploreView } from "./views/ExploreView";

const renderView = (state: State) => {
  switch (state.view) {
    case "login": return AuthForm({ isLogin: true })(state);
    case "signup": return AuthForm({ isLogin: false })(state);
    case "profile": return ProfileView(state);
    case "search": return UserSearchView(state);
    case "leaderboard": return LeaderboardView(state);
    case "explore": return ExploreView(state);
    case "not_found":
    default:
      return h<State>("div", { class: "w-full h-screen flex items-center justify-center" }, text("404 - Page Not Found"));
  }
}

const view = (state: State) => {
  const viewsWithoutSidebar: State["view"][] = ["login", "signup"];
  const page = renderView(state);

  if (viewsWithoutSidebar.includes(state.view)) {
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
