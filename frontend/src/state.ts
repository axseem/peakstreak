// frontend/src/state.ts

import { api } from "./api";
import type { State, User, Habit } from "./types";
import { path_to_view, NavigateFx } from "./router"; // Import router helpers

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

export const SetView = (state: State, view: State["view"]): State => ({ ...state, view, error: null });

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

export const SetHabits = (state: State, habits: Habit[]): State => {
  const habitsWithStatus = habits.map(h => ({ ...h, loggedToday: false }));
  return { ...state, habits: habitsWithStatus, isLoading: false };
};

export const MarkHabitAsLogged = (state: State, habitId: string): State => ({
  ...state,
  isLoading: false,
  habits: state.habits.map(h => h.id === habitId ? { ...h, loggedToday: true } : h),
});

export const Logout = (_state: State): [State, any] => {
  localStorage.removeItem("peakstreak_user");
  localStorage.removeItem("peakstreak_token");
  const newState: State = { ...initialState, view: "login", user: null, token: null };
  return [newState, [NavigateFx, { path: "/login", replace: true }]];
};


// --- Effects (Asynchronous Side-Effects) ---

export const FetchHabitsFx = (dispatch: any, { token, username }: { token: string, username: string }) => {
  dispatch(SetLoading, true);
  api.get(`/api/user/${username}`, token)
    .then(data => dispatch(SetHabits, data.habits || []))
    .catch(err => dispatch(SetError, err.message));
};

export const LoginFx = (dispatch: any, { identifier, password }: any) => {
  dispatch(SetLoading, true);
  api.post("/api/auth/login", { identifier, password })
    .then(data => {
      // SetAuth will handle the state update and navigation
      dispatch(SetAuth, data);
      FetchHabitsFx(dispatch, { token: data.token, username: data.user.username });
    })
    .catch(err => dispatch(SetError, err.message));
};

export const SignUpFx = (dispatch: any, { username, email, password }: any) => {
  dispatch(SetLoading, true);
  api.post("/api/auth/signup", { username, email, password })
    .then(() => dispatch(NavigateFx, { path: "/login" }))
    .catch(err => dispatch(SetError, err.message));
};

export const CreateHabitFx = (dispatch: any, { name, token, username }: { name: string, token: string, username: string }) => {
  dispatch(SetLoading, true);
  api.post("/api/habit", { name }, token)
    .then(() => {
      FetchHabitsFx(dispatch, { token, username });
    })
    .catch(err => dispatch(SetError, err.message));
};

export const LogHabitFx = (dispatch: any, { habitId, token }: { habitId: string, token: string }) => {
  const date = new Date().toISOString().split('T')[0];
  dispatch(SetLoading, true);
  api.post(`/api/habit/${habitId}/log`, { date, status: true }, token)
    .then(() => dispatch(MarkHabitAsLogged, habitId))
    .catch(err => dispatch(SetError, err.message));
};

export const initFx = (dispatch: any, state: State) => {
  if (state.view === "dashboard" && (!state.user || !state.token)) {
    dispatch(NavigateFx, { path: "/login", replace: true });
    return;
  }

  if (state.user && state.token) {
    FetchHabitsFx(dispatch, { token: state.token, username: state.user.username });
  }
};
