import type { State } from "./types";
import { SetView } from "./state";

const routes: Record<string, State["view"]> = {
  "/login": "login",
  "/signup": "signup",
  "/dashboard": "dashboard",
};

export const path_to_view = (path: string): State["view"] => {
  return routes[path] || "/dashboard";
};

export const NavigateFx = (dispatch: any, { path, replace = false }: { path: string, replace?: boolean }) => {
  const newView = path_to_view(path);
  if (replace) {
    history.replaceState({ view: newView }, "", path);
  } else {
    history.pushState({ view: newView }, "", path);
  }
  dispatch(SetView, newView);
};

const onPopState = (dispatch: any) => {
  const handler = () => {
    const newView = path_to_view(window.location.pathname);
    dispatch(SetView, newView);
  };
  window.addEventListener("popstate", handler);
  return () => window.removeEventListener("popstate", handler);
};

export const PopStateSub = (dispatch: any) => onPopState(dispatch);
