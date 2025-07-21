import type { State } from "./types";
import { SetView } from "./state";

const staticRoutes: Record<string, State["view"]> = {
  "/login": "login",
  "/signup": "signup",
  "/leaderboard": "leaderboard",
  "/explore": "explore",
};

export const path_to_view = (path: string): { view: State["view"], username?: string } => {
  if (path === "/") {
    return { view: "home" };
  }
  if (staticRoutes[path]) {
    return { view: staticRoutes[path] };
  }
  if (path.startsWith("/@")) {
    const username = path.substring(2);
    if (username) {
      return { view: "profile", username };
    }
  }
  return { view: "not_found" };
};

export const NavigateFx = (dispatch: any, { path, replace = false }: { path: string, replace?: boolean }) => {
  const { view, username } = path_to_view(path);
  if (replace) {
    history.replaceState({ view, username }, "", path);
  } else {
    history.pushState({ view, username }, "", path);
  }
  dispatch(SetView, { view, username });
};

const onPopState = (dispatch: any) => {
  const handler = (event: PopStateEvent) => {
    const { view, username } = event.state || path_to_view(window.location.pathname);
    dispatch(SetView, { view, username });
  };
  window.addEventListener("popstate", handler);
  return () => window.removeEventListener("popstate", handler);
};

export const PopStateSub = (dispatch: any) => onPopState(dispatch);
