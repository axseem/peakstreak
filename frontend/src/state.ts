import { api } from "./api";
import type { State, User, HabitWithLogs, HabitLog, ProfileData } from "./types";
import { path_to_view, NavigateFx } from "./router";

const savedUser = localStorage.getItem("peakstreak_user");
const savedToken = localStorage.getItem("peakstreak_token");

const { view: initialView } = path_to_view(window.location.pathname);

export const initialState: State = {
  view: initialView,
  user: savedUser ? JSON.parse(savedUser) : null,
  token: savedToken,
  profileData: null,
  isLoading: false,
  error: null,
  newHabitName: "",
  isAddingHabit: false,
  isProfileMenuOpen: false,
};

// --- Actions (Synchronous State Updaters) ---

export const SetView = (state: State, { view, username }: { view: State["view"], username?: string }): [State, any] | State => {
  if (view === "home") {
    if (state.token && state.user) {
      return [state, [NavigateFx, { path: `/@${state.user.username}`, replace: true }]];
    }
    return [state, [NavigateFx, { path: "/login", replace: true }]];
  }

  const authRequiredViews: State["view"][] = [];
  if (authRequiredViews.includes(view) && !state.token) {
    return [state, [NavigateFx, { path: "/login", replace: true }]];
  }

  const newState = { ...state, view, error: null, isLoading: true };

  if (view === "profile" && username) {
    return [newState, [FetchProfileFx, { username, token: state.token }]];
  }

  return { ...newState, isLoading: false };
};

export const SetError = (state: State, error: string | null): State => {
  console.error("error: " + error);
  return { ...state, isLoading: false, error, newHabitName: "", isAddingHabit: false };
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
  return [newState, [NavigateFx, { path: `/@${user.username}`, replace: true }]];
};

export const SetProfileData = (state: State, profileData: ProfileData): State => {
  return { ...state, profileData, isLoading: false };
};

export const AddHabit = (state: State, newHabit: HabitWithLogs): State => {
  if (!state.profileData) return state;
  return {
    ...state,
    isLoading: false,
    newHabitName: "",
    isAddingHabit: false,
    profileData: {
      ...state.profileData,
      habits: [newHabit, ...state.profileData.habits]
    }
  };
};

export const AddHabitLog = (state: State, { habitId, log }: { habitId: string, log: HabitLog }): State => {
  if (!state.profileData) return state;
  return {
    ...state,
    isLoading: false,
    profileData: {
      ...state.profileData,
      habits: state.profileData.habits.map(h =>
        h.id === habitId
          ? { ...h, logs: [...h.logs, log].sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()) }
          : h
      ),
    }
  };
};

export const Logout = (_state: State): [State, any] => {
  localStorage.removeItem("peakstreak_user");
  localStorage.removeItem("peakstreak_token");
  const newState: State = { ...initialState, view: "login", user: null, token: null, profileData: null, isProfileMenuOpen: false };
  return [newState, [NavigateFx, { path: "/login", replace: true }]];
};

export const ShowAddHabitForm = (state: State): State => ({
  ...state,
  isAddingHabit: true,
});

export const HideAddHabitForm = (state: State): State => ({
  ...state,
  isAddingHabit: false,
  newHabitName: "",
});

export const ToggleProfileMenu = (state: State): State => ({
  ...state,
  isProfileMenuOpen: !state.isProfileMenuOpen,
});

export const CloseProfileMenu = (state: State): State => ({
  ...state,
  isProfileMenuOpen: false,
});


// --- Effects (Asynchronous Side-Effects) ---

export const FetchProfileFx = (dispatch: any, { username, token }: { username: string, token: string | null }) => {
  dispatch(SetLoading, true);
  api.get(`/api/profile/${username}`, token)
    .then(data => dispatch(SetProfileData, data))
    .catch(err => dispatch(SetError, err.message));
};

export const LoginFx = (dispatch: any, { identifier, password }: any) => {
  dispatch(SetLoading, true);
  api.post("/api/auth/login", { identifier, password })
    .then(data => dispatch(SetAuth, data))
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
    .then((newHabit) => {
      const newHabitWithLogs: HabitWithLogs = { ...newHabit, logs: [] };
      dispatch(AddHabit, newHabitWithLogs);
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

export const initFx = (dispatch: any, _state: State) => {
  const { view, username } = path_to_view(window.location.pathname);
  dispatch(SetView, { view, username });
};
