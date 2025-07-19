import { api } from "./api";
import type { State, User, Habit } from "./types";

// Initial State
const savedUser = localStorage.getItem("peakstreak_user");
const savedToken = localStorage.getItem("peakstreak_token");

export const initialState: State = {
  view: savedUser && savedToken ? "dashboard" : "login",
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
  console.log("log of error:" + error);
  return { ...state, isLoading: false, error };
};

export const SetLoading = (state: State, isLoading: boolean): State => ({ ...state, isLoading });

export const HandleFormInput = (state: State, event: Event): State => ({
  ...state,
  newHabitName: (event.target as HTMLInputElement).value,
});

export const SetAuth = (state: State, { user, token }: { user: User, token: string }): State => {
  localStorage.setItem("peakstreak_user", JSON.stringify(user));
  localStorage.setItem("peakstreak_token", token);
  return { ...state, user, token, view: "dashboard", isLoading: false, error: null };
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

export const Logout = (_state: State): State => {
  localStorage.removeItem("peakstreak_user");
  localStorage.removeItem("peakstreak_token");
  return { ...initialState, view: "login" };
};


// --- Effects (Asynchronous Side-Effects) ---

export const FetchHabitsFx = (dispatch: any, { token, username }: { token: string, username: string }) => {
  dispatch(SetLoading, true);
  api.get(`/api/users/${username}`, token)
    .then(data => dispatch(SetHabits, data.habits || []))
    .catch(err => dispatch(SetError, err.message));
};

export const LoginFx = (dispatch: any, { identifier, password }: any) => {
  dispatch(SetLoading, true);
  api.post("/api/auth/login", { identifier, password })
    .then(data => {
      dispatch(SetAuth, data);
      FetchHabitsFx(dispatch, { token: data.token, username: data.user.username });
    })
    .catch(err => dispatch(SetError, err.message));
};

export const SignUpFx = (dispatch: any, { username, email, password }: any) => {
  dispatch(SetLoading, true);
  api.post("/api/auth/signup", { username, email, password })
    .then(() => dispatch(SetView, "login"))
    .catch(err => dispatch(SetError, err.message));
};

export const CreateHabitFx = (dispatch: any, { name, token, username }: { name: string, token: string, username: string }) => {
  dispatch(SetLoading, true);
  api.post("/api/habits", { name }, token)
    .then(() => {
      FetchHabitsFx(dispatch, { token, username });
    })
    .catch(err => dispatch(SetError, err.message));
};

export const LogHabitFx = (dispatch: any, { habitId, token }: { habitId: string, token: string }) => {
  const date = new Date().toISOString().split('T')[0];
  dispatch(SetLoading, true);
  api.post(`/api/habits/${habitId}/logs`, { date, status: true }, token)
    .then(() => dispatch(MarkHabitAsLogged, habitId))
    .catch(err => dispatch(SetError, err.message));
};

export const initFx = (dispatch: any, state: State) => {
  if (state.user && state.token) {
    FetchHabitsFx(dispatch, { token: state.token, username: state.user.username });
  }
}
