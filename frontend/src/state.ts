import { api } from "./api";
import type { State, User, HabitWithLogs, HabitLog } from "./types";
import { path_to_view, NavigateFx } from "./router";

const savedUser = localStorage.getItem("peakstreak_user");
const savedToken = localStorage.getItem("peakstreak_token");

const initialPath = window.location.pathname;
const initialView = path_to_view(initialPath);

export const initialState: State = {
  view: initialView,
  user: savedUser ? JSON.parse(savedUser) : null,
  token: savedToken,
  habits: [],
  isLoading: false,
  error: null,
  newHabitName: "",
};


// --- Actions (Synchronous State Updaters) ---

export const SetView = (state: State, view: State["view"]): [State, any] | State => {
  if (state.view === view && view !== "dashboard") {
    return state;
  }

  const authRequired = view === "dashboard";
  if (authRequired && !state.token) {
    return [state, [NavigateFx, { path: "/login", replace: true }]];
  }

  const newState = { ...state, view, error: null };

  if (view === "dashboard" && newState.token) {
    return [newState, [FetchHabitsFx, { token: newState.token }]];
  }

  return newState;
};

export const SetError = (state: State, error: string | null): State => {
  console.error("error: " + error);
  return { ...state, isLoading: false, error };
};

export const SetLoading = (state: State, isLoading: boolean): State => ({ ...state, isLoading });

export const HandleFormInput = (state: State, event: Event): State => ({
  ...state,
  newHabitName: (event.target as HTMLInputElement).value,
});

export const SetAuth = (state: State, { user, token }: { user: User, token: string }): [State, any] => {
  localStorage.setItem("peakstreak_user", JSON.stringify(user));
  localStorage.setItem("peakstreak_token", token);
  const newState = { ...state, user, token, isLoading: false, error: null };
  return [newState, [NavigateFx, { path: "/dashboard", replace: true }]];
};

export const SetHabits = (state: State, habits: HabitWithLogs[]): State => {
  return { ...state, habits, isLoading: false };
};

export const AddHabitLog = (state: State, { habitId, log }: { habitId: string, log: HabitLog }): State => ({
  ...state,
  isLoading: false,
  habits: state.habits.map(h =>
    h.id === habitId
      ? { ...h, logs: [...h.logs, log].sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()) }
      : h
  ),
});

export const Logout = (_state: State): [State, any] => {
  localStorage.removeItem("peakstreak_user");
  localStorage.removeItem("peakstreak_token");
  const newState: State = { ...initialState, view: "login", user: null, token: null };
  return [newState, [NavigateFx, { path: "/login", replace: true }]];
};


// --- Effects (Asynchronous Side-Effects) ---

export const FetchHabitsFx = (dispatch: any, { token }: { token: string }) => {
  dispatch(SetLoading, true);
  api.get(`/api/me/habit`, token)
    .then(data => dispatch(SetHabits, data || []))
    .catch(err => dispatch(SetError, err.message));
};

export const LoginFx = (dispatch: any, { identifier, password }: any) => {
  dispatch(SetLoading, true);
  api.post("/api/auth/login", { identifier, password })
    .then(data => {
      dispatch(SetAuth, data);
    })
    .catch(err => dispatch(SetError, err.message));
};

export const SignUpFx = (dispatch: any, { username, email, password }: any) => {
  dispatch(SetLoading, true);
  api.post("/api/auth/signup", { username, email, password })
    .then(() => dispatch(NavigateFx, { path: "/login" }))
    .catch(err => dispatch(SetError, err.message));
};

export const CreateHabitFx = (dispatch: any, { name, token }: { name: string, token: string }) => {
  dispatch(SetLoading, true);
  api.post("/api/habit", { name }, token)
    .then(() => {
      // Re-trigger SetView to refresh habits
      dispatch(SetView, "dashboard");
    })
    .catch(err => dispatch(SetError, err.message));
};

export const LogHabitFx = (dispatch: any, { habitId, token }: { habitId: string, token: string }) => {
  const date = new Date().toISOString().split('T')[0];
  dispatch(SetLoading, true);
  api.post(`/api/habit/${habitId}/log`, { date, status: true }, token)
    .then((newLog) => dispatch(AddHabitLog, { habitId, log: newLog }))
    .catch(err => dispatch(SetError, err.message));
};

export const initFx = (dispatch: any, state: State) => {
  if (window.location.pathname === "/") {
    history.replaceState({ view: 'dashboard' }, "", '/dashboard');
  }

  dispatch(SetView, state.view);
};
