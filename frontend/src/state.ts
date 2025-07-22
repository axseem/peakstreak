import { api } from "./api";
import type { State, User, HabitWithLogs, HabitLog, ProfileData } from "./types";
import { path_to_view, NavigateFx } from "./router";
import { toYYYYMMDD } from "./lib/date";

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
  isEditing: false,
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

  const newState = { ...state, view, error: null, isLoading: true, isEditing: false };

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

export const UpdateHabitLog = (state: State, { habitId, log }: { habitId: string, log: HabitLog }): State => {
  if (!state.profileData) return state;

  return {
    ...state,
    isLoading: false,
    profileData: {
      ...state.profileData,
      habits: state.profileData.habits.map(h => {
        if (h.id !== habitId) return h;

        const logDate = toYYYYMMDD(new Date(log.date));
        const existingLogIndex = h.logs.findIndex(l => toYYYYMMDD(new Date(l.date)) === logDate);
        let newLogs;

        if (existingLogIndex > -1) {
          if (log.status) {
            newLogs = [...h.logs];
            newLogs[existingLogIndex] = log;
          } else {
            newLogs = h.logs.filter(l => toYYYYMMDD(new Date(l.date)) !== logDate);
          }
        } else if (log.status) {
          newLogs = [...h.logs, log];
        } else {
          newLogs = h.logs;
        }

        return { ...h, logs: newLogs.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()) };
      })
    }
  };
};

export const UpdateHabitName = (state: State, { habitId, name }: { habitId: string, name: string }): State => {
  if (!state.profileData) return state;
  return {
    ...state,
    isLoading: false,
    profileData: {
      ...state.profileData,
      habits: state.profileData.habits.map(h =>
        h.id === habitId ? { ...h, name } : h
      )
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

export const ToggleEditMode = (state: State): State => ({
  ...state,
  isEditing: !state.isEditing,
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
    .then((newLog) => dispatch(UpdateHabitLog, { habitId, log: newLog }))
    .catch(err => dispatch(SetError, err.message));
};

export const UpdateHabitNameFx = (dispatch: any, { habitId, name, token }: { habitId: string, name: string, token: string }) => {
  dispatch(SetLoading, true);
  api.put(`/api/habit/${habitId}`, { name }, token)
    .then(() => dispatch(UpdateHabitName, { habitId, name }))
    .catch(err => dispatch(SetError, err.message));
};

export const ToggleHabitLogFx = (dispatch: any, { habitId, date, currentStatus, token }: { habitId: string, date: string, currentStatus: boolean, token: string }) => {
  dispatch(SetLoading, true);
  const newStatus = !currentStatus;
  api.post(`/api/habit/${habitId}/log`, { date, status: newStatus }, token)
    .then((newLog) => dispatch(UpdateHabitLog, { habitId, log: newLog }))
    .catch(err => dispatch(SetError, err.message));
};

export const initFx = (dispatch: any, _state: State) => {
  const { view, username } = path_to_view(window.location.pathname);
  dispatch(SetView, { view, username });
};
