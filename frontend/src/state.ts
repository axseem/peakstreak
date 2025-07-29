import { api } from "./api";
import type {
  State,
  User,
  Habit,
  HabitWithLogs,
  HabitLog,
  ProfileData,
  PublicUser,
  SearchState,
  LeaderboardEntry,
  ExploreEntry,
} from "./types";
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
  newHabitColorHue: 180,
  newHabitIsBoolean: true,
  isAddingHabit: false,
  isProfileMenuOpen: false,
  editingHabit: null,
  activeHabitMenuId: null,
  followerList: {
    isOpen: false,
    type: null,
    users: [],
    isLoading: false,
    error: null,
    title: "",
  },
  search: {
    query: "",
    results: [],
    isLoading: false,
    error: null,
  },
  leaderboard: {
    users: [],
    isLoading: false,
    error: null,
  },
  explore: {
    entries: [],
    isLoading: false,
    error: null,
  },
};

// --- Actions (Synchronous State Updaters) ---

export const SetView = (
  state: State,
  { view, username }: { view: State["view"]; username?: string },
): [State, any] | State => {
  const baseState = {
    ...state,
    view,
    error: null,
    editingHabit: null,
    activeHabitMenuId: null,
  };

  if (view === "home") {
    if (state.token && state.user) {
      return [
        baseState,
        [NavigateFx, { path: `/@${state.user.username}`, replace: true }],
      ];
    }
    return [baseState, [NavigateFx, { path: "/login", replace: true }]];
  }

  const authRequiredViews: State["view"][] = ["settings"];
  if (authRequiredViews.includes(view) && !state.token) {
    return [baseState, [NavigateFx, { path: "/login", replace: true }]];
  }

  if (view === "explore") {
    const newState = {
      ...baseState,
      explore: { ...state.explore, isLoading: true, error: null },
    };
    return [newState, [FetchExploreDataFx, {}]];
  }

  if (view === "leaderboard") {
    const newState = {
      ...baseState,
      leaderboard: { ...state.leaderboard, isLoading: true, error: null },
    };
    return [newState, [FetchLeaderboardFx, {}]];
  }

  if (view === "search") {
    const urlParams = new URLSearchParams(window.location.search);
    const query = urlParams.get("q") || "";
    const searchState: SearchState = {
      query,
      isLoading: true,
      error: null,
      results: [],
    };
    const newState = { ...baseState, search: searchState };

    if (query.length >= 3) {
      return [newState, [SearchUsersFx, { query }]];
    }
    return { ...newState, search: { ...searchState, isLoading: false } };
  }

  const newState = { ...baseState, isLoading: true };

  if (view === "profile" && username) {
    return [newState, [FetchProfileFx, { username, token: state.token }]];
  }

  return { ...newState, isLoading: false };
};

export const SetError = (state: State, error: string | null): State => {
  console.error("error: " + error);
  return {
    ...state,
    isLoading: false,
    error,
    newHabitName: "",
    isAddingHabit: false,
  };
};

export const SetLoading = (state: State, isLoading: boolean): State => ({
  ...state,
  isLoading,
});

export const HandleFormInput = (state: State, event: Event): State => ({
  ...state,
  newHabitName: (event.target as HTMLInputElement).value,
});

export const HandleColorInput = (state: State, event: Event): State => ({
  ...state,
  newHabitColorHue: parseInt((event.target as HTMLInputElement).value, 10),
});

export const HandleIsBooleanInput = (state: State, event: Event): State => ({
  ...state,
  newHabitIsBoolean: (event.target as HTMLInputElement).checked,
});

export const SetAuth = (
  state: State,
  { user, token }: { user: User; token: string },
): [State, any] => {
  localStorage.setItem("peakstreak_user", JSON.stringify(user));
  localStorage.setItem("peakstreak_token", token);
  const newState = { ...state, user, token, isLoading: false, error: null };
  return [
    newState,
    [NavigateFx, { path: `/@${user.username}`, replace: true }],
  ];
};

export const SetProfileData = (
  state: State,
  profileData: ProfileData,
): State => {
  return { ...state, profileData, isLoading: false };
};

export const SetAvatarUrl = (
  state: State,
  { avatarUrl }: { avatarUrl: string },
): State => {
  if (!state.user || !state.profileData) return state;

  const updatedUser = { ...state.user, avatarUrl };
  const updatedProfileData = {
    ...state.profileData,
    user: { ...state.profileData.user, avatarUrl },
  };

  localStorage.setItem("peakstreak_user", JSON.stringify(updatedUser));

  return {
    ...state,
    user: updatedUser,
    profileData: updatedProfileData,
    isLoading: false,
    error: null,
  };
};

export const AddHabit = (state: State, newHabit: HabitWithLogs): State => {
  if (!state.profileData) return state;
  return {
    ...state,
    isLoading: false,
    newHabitName: "",
    newHabitColorHue: 180,
    newHabitIsBoolean: true,
    isAddingHabit: false,
    profileData: {
      ...state.profileData,
      habits: [...state.profileData.habits, newHabit],
    },
  };
};

export const UpdateHabitLog = (
  state: State,
  { habitId, log }: { habitId: string; log: HabitLog },
): State => {
  if (!state.profileData) return state;

  return {
    ...state,
    isLoading: false,
    profileData: {
      ...state.profileData,
      habits: state.profileData.habits.map((h) => {
        if (h.id !== habitId) return h;

        const logDate = toYYYYMMDD(new Date(log.date));
        const existingLogIndex = h.logs.findIndex(
          (l) => toYYYYMMDD(new Date(l.date)) === logDate,
        );
        let newLogs;

        if (existingLogIndex > -1) {
          if (log.value > 0) {
            newLogs = [...h.logs];
            newLogs[existingLogIndex] = log;
          } else {
            newLogs = h.logs.filter(
              (l) => toYYYYMMDD(new Date(l.date)) !== logDate,
            );
          }
        } else if (log.value > 0) {
          newLogs = [...h.logs, log];
        } else {
          newLogs = h.logs;
        }

        return {
          ...h,
          logs: newLogs.sort(
            (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime(),
          ),
        };
      }),
    },
  };
};

export const UpdateHabit = (
  state: State,
  {
    habitId,
    name,
    colorHue,
  }: { habitId: string; name: string; colorHue: number },
): State => {
  if (!state.profileData) return state;
  return {
    ...state,
    isLoading: false,
    editingHabit: null,
    profileData: {
      ...state.profileData,
      habits: state.profileData.habits.map((h) =>
        h.id === habitId ? { ...h, name, colorHue } : h,
      ),
    },
  };
};

export const HabitDeleted = (state: State, habitId: string): State => {
  if (!state.profileData) return state;
  return {
    ...state,
    isLoading: false,
    profileData: {
      ...state.profileData,
      habits: state.profileData.habits.filter((h) => h.id !== habitId),
    },
  };
};

export const Logout = (_state: State): [State, any] => {
  localStorage.removeItem("peakstreak_user");
  localStorage.removeItem("peakstreak_token");
  const newState: State = {
    ...initialState,
    view: "login",
    user: null,
    token: null,
    profileData: null,
    isProfileMenuOpen: false,
  };
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
  newHabitColorHue: 180,
  newHabitIsBoolean: true,
});

export const ToggleProfileMenu = (state: State): State => ({
  ...state,
  isProfileMenuOpen: !state.isProfileMenuOpen,
});

export const CloseProfileMenu = (state: State): State => ({
  ...state,
  isProfileMenuOpen: false,
});

export const ToggleHabitMenu = (state: State, habitId: string): State => ({
  ...state,
  activeHabitMenuId: state.activeHabitMenuId === habitId ? null : habitId,
});

export const CloseHabitMenu = (state: State): State => ({
  ...state,
  activeHabitMenuId: null,
});

export const StartEditingHabit = (state: State, habit: Habit): State => ({
  ...state,
  editingHabit: {
    id: habit.id,
    name: habit.name,
    colorHue: habit.colorHue,
  },
  activeHabitMenuId: null,
});

export const CancelEditingHabit = (state: State): State => ({
  ...state,
  editingHabit: null,
});

export const UpdateEditingHabitField = (
  state: State,
  { field, value }: { field: "name" | "colorHue"; value: string | number },
): State => {
  if (!state.editingHabit) return state;
  return {
    ...state,
    editingHabit: { ...state.editingHabit, [field]: value },
  };
};

export const SetFollowingStatus = (
  state: State,
  { isFollowing }: { isFollowing: boolean },
): State => {
  if (!state.profileData) return state;
  const currentFollowers = state.profileData.followersCount;
  return {
    ...state,
    isLoading: false,
    profileData: {
      ...state.profileData,
      isFollowing,
      followersCount: isFollowing ? currentFollowers + 1 : currentFollowers - 1,
    },
  };
};

export const OpenFollowerList = (
  state: State,
  { type, username }: { type: "followers" | "following"; username: string },
): [State, any] => {
  const title = type === "followers" ? "Followers" : "Following";
  const effect = type === "followers" ? FetchFollowersFx : FetchFollowingFx;
  const newState = {
    ...state,
    followerList: {
      isOpen: true,
      type,
      title,
      isLoading: true,
      error: null,
      users: [],
    },
  };
  return [newState, [effect, { username, token: state.token }]];
};

export const CloseFollowerList = (state: State): State => ({
  ...state,
  followerList: {
    ...state.followerList,
    isOpen: false,
  },
});

export const SetFollowerListData = (
  state: State,
  { users }: { users: PublicUser[] },
): State => ({
  ...state,
  followerList: {
    ...state.followerList,
    users,
    isLoading: false,
    error: null,
  },
});

export const SetFollowerListError = (state: State, error: string): State => ({
  ...state,
  followerList: {
    ...state.followerList,
    isLoading: false,
    error,
  },
});

export const SetSearchResults = (
  state: State,
  { query, results }: { query: string; results: PublicUser[] },
): State => ({
  ...state,
  search: {
    query,
    results,
    isLoading: false,
    error: null,
  },
});

export const SetSearchError = (
  state: State,
  { query, error }: { query: string; error: string },
): State => ({
  ...state,
  search: {
    query,
    results: [],
    isLoading: false,
    error,
  },
});

export const SetLeaderboardData = (
  state: State,
  { users }: { users: LeaderboardEntry[] },
): State => ({
  ...state,
  leaderboard: {
    users,
    isLoading: false,
    error: null,
  },
});

export const SetLeaderboardError = (state: State, error: string): State => ({
  ...state,
  leaderboard: {
    ...state.leaderboard,
    users: [],
    isLoading: false,
    error,
  },
});

export const SetExploreData = (
  state: State,
  { entries }: { entries: ExploreEntry[] },
): State => ({
  ...state,
  explore: {
    entries,
    isLoading: false,
    error: null,
  },
});

export const SetExploreError = (state: State, error: string): State => ({
  ...state,
  explore: {
    ...state.explore,
    entries: [],
    isLoading: false,
    error,
  },
});

// --- Effects (Asynchronous Side-Effects) ---

export const FetchProfileFx = (
  dispatch: any,
  { username, token }: { username: string; token: string | null },
) => {
  dispatch(SetLoading, true);
  api
    .get(`/api/profile/${username}`, token)
    .then((data) => dispatch(SetProfileData, data))
    .catch((err) => dispatch(SetError, err.message));
};

export const LoginFx = (dispatch: any, { identifier, password }: any) => {
  dispatch(SetLoading, true);
  api
    .post("/api/auth/login", { identifier, password })
    .then((data) => dispatch(SetAuth, data))
    .catch((err) => dispatch(SetError, err.message));
};

export const SignUpFx = (dispatch: any, { username, email, password }: any) => {
  dispatch(SetLoading, true);
  api
    .post("/api/auth/signup", { username, email, password })
    .then(() => {
      const GoToLogin = (state: State): [State, any] => [
        state,
        [NavigateFx, { path: "/login" }],
      ];
      dispatch(GoToLogin);
    })
    .catch((err) => dispatch(SetError, err.message));
};

export const CreateHabitFx = (
  dispatch: any,
  {
    name,
    colorHue,
    isBoolean,
    token,
  }: { name: string; colorHue: number; isBoolean: boolean; token: string },
) => {
  dispatch(SetLoading, true);
  api
    .post("/api/habit", { name, colorHue, isBoolean }, token)
    .then((newHabit) => {
      const newHabitWithLogs: HabitWithLogs = { ...newHabit, logs: [] };
      dispatch(AddHabit, newHabitWithLogs);
    })
    .catch((err) => dispatch(SetError, err.message));
};

export const UpsertHabitLogFx = (
  dispatch: any,
  {
    habitId,
    date,
    value,
    token,
  }: { habitId: string; date: string; value: number; token: string },
) => {
  dispatch(SetLoading, true);
  api
    .post(`/api/habit/${habitId}/log`, { date, value }, token)
    .then((newLog) => dispatch(UpdateHabitLog, { habitId, log: newLog }))
    .catch((err) => dispatch(SetError, err.message));
};

export const UpdateHabitFx = (
  dispatch: any,
  {
    habitId,
    name,
    colorHue,
    token,
  }: { habitId: string; name: string; colorHue: number; token: string },
) => {
  dispatch(SetLoading, true);
  api
    .put(`/api/habit/${habitId}`, { name, colorHue }, token)
    .then(() => dispatch(UpdateHabit, { habitId, name, colorHue }))
    .catch((err) => dispatch(SetError, err.message));
};

export const DeleteHabitFx = (
  dispatch: any,
  { habitId, token }: { habitId: string; token: string },
) => {
  dispatch(SetLoading, true);
  api
    .delete(`/api/habit/${habitId}`, token)
    .then(() => dispatch(HabitDeleted, habitId))
    .catch((err) => dispatch(SetError, err.message));
};

export const FollowUserFx = (
  dispatch: any,
  { username, token }: { username: string; token: string },
) => {
  dispatch(SetLoading, true);
  api
    .post(`/api/profile/${username}/follow`, null, token)
    .then(() => dispatch(SetFollowingStatus, { isFollowing: true }))
    .catch((err) => dispatch(SetError, err.message));
};

export const UnfollowUserFx = (
  dispatch: any,
  { username, token }: { username: string; token: string },
) => {
  dispatch(SetLoading, true);
  api
    .delete(`/api/profile/${username}/follow`, token)
    .then(() => dispatch(SetFollowingStatus, { isFollowing: false }))
    .catch((err) => dispatch(SetError, err.message));
};

export const FetchFollowersFx = (
  dispatch: any,
  { username, token }: { username: string; token: string | null },
) => {
  api
    .get(`/api/profile/${username}/followers`, token)
    .then((users) => dispatch(SetFollowerListData, { users }))
    .catch((err) => dispatch(SetFollowerListError, err.message));
};

export const FetchFollowingFx = (
  dispatch: any,
  { username, token }: { username: string; token: string | null },
) => {
  api
    .get(`/api/profile/${username}/following`, token)
    .then((users) => dispatch(SetFollowerListData, { users }))
    .catch((err) => dispatch(SetFollowerListError, err.message));
};

const KB = 1024;
const MB = 1024 * KB;

export const UploadAvatarFx = (
  dispatch: any,
  { file, token }: { file: File; token: string },
) => {
  if (file.size > 2 * MB) {
    dispatch(SetError, "File is too large. Max 2MB.");
    return;
  }
  dispatch(SetLoading, true);
  const formData = new FormData();
  formData.append("avatar", file);

  api
    .upload("/api/user/avatar", formData, token)
    .then((data) => dispatch(SetAvatarUrl, { avatarUrl: data.avatarUrl }))
    .catch((err) => dispatch(SetError, err.message));
};

export const SearchUsersFx = (dispatch: any, { query }: { query: string }) => {
  api
    .get(`/api/users/search?q=${encodeURIComponent(query)}`, null)
    .then((results) => dispatch(SetSearchResults, { query, results }))
    .catch((err) => dispatch(SetSearchError, { query, error: err.message }));
};

export const FetchLeaderboardFx = (dispatch: any) => {
  api
    .get("/api/leaderboard", null)
    .then((data) => dispatch(SetLeaderboardData, { users: data }))
    .catch((err) => dispatch(SetLeaderboardError, err.message));
};

export const FetchExploreDataFx = (dispatch: any) => {
  api
    .get("/api/explore", null)
    .then((data) => dispatch(SetExploreData, { entries: data }))
    .catch((err) => dispatch(SetExploreError, err.message));
};

export const DeleteAccountFx = (
  dispatch: any,
  { token }: { token: string },
) => {
  api
    .delete("/api/user", token)
    .then(() => {
      dispatch(Logout);
    })
    .catch((err) =>
      dispatch(SetError, `Failed to delete account: ${err.message}`),
    );
};

export const initFx = (dispatch: any, _state: State) => {
  const { view, username } = path_to_view(window.location.pathname);
  dispatch(SetView, { view, username });
};
