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

export type State = {
  view: "login" | "signup" | "profile" | "leaderboard" | "explore" | "not_found" | "home";
  user: User | null;
  token: string | null;
  profileData: ProfileData | null;
  isLoading: boolean;
  error: string | null;
  newHabitName: string;
  isAddingHabit: boolean;
  isProfileMenuOpen: boolean;
  isEditing: boolean;
  followerList: FollowerListState;
};
