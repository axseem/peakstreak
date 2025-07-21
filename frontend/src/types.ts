export type User = {
  id: string;
  username: string;
  email: string;
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
};

export type State = {
  view: "login" | "signup" | "profile" | "leaderboard" | "explore" | "not_found" | "home";
  user: User | null;
  token: string | null;
  profileData: ProfileData | null;
  isLoading: boolean;
  error: string | null;
  newHabitName: string;
};
