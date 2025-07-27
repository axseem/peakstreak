export type User = {
  id: string;
  username: string;
  email: string;
  avatarUrl?: string;
};

export type PublicUser = {
  id: string;
  username: string;
  avatarUrl?: string;
};

export type Habit = {
  id: string;
  userId: string;
  name: string;
  colorHue: number;
  createdAt: string;
};

export type HabitLog = {
  id: string;
  habitId: string;
  date: string;
  status: boolean;
};

export type HabitWithLogs = Habit & {
  logs: HabitLog[];
};

export type ProfileData = {
  user: User;
  habits: HabitWithLogs[];
  isOwner: boolean;
  followersCount: number;
  followingCount: number;
  isFollowing: boolean;
};

export type FollowerListState = {
  isOpen: boolean;
  type: 'followers' | 'following' | null;
  users: PublicUser[];
  isLoading: boolean;
  error: string | null;
  title: string;
};

export type SearchState = {
  query: string;
  results: PublicUser[];
  isLoading: boolean;
  error: string | null;
};

export type LeaderboardEntry = {
  user: PublicUser;
  totalLoggedDays: number;
  habits: HabitWithLogs[];
};

export type LeaderboardState = {
  users: LeaderboardEntry[];
  isLoading: boolean;
  error: string | null;
};

export type ExploreEntry = {
  user: PublicUser;
  habit: HabitWithLogs;
};

export type ExploreState = {
  entries: ExploreEntry[];
  isLoading: boolean;
  error: string | null;
};

export type State = {
  view: "login" | "signup" | "profile" | "leaderboard" | "explore" | "not_found" | "home" | "search";
  user: User | null;
  token: string | null;
  profileData: ProfileData | null;
  isLoading: boolean;
  error: string | null;
  newHabitName: string;
  newHabitColorHue: number;
  isAddingHabit: boolean;
  isProfileMenuOpen: boolean;
  editingHabitId: string | null;
  activeHabitMenuId: string | null;
  followerList: FollowerListState;
  search: SearchState;
  leaderboard: LeaderboardState;
  explore: ExploreState;
};
