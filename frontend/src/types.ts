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

export type State = {
  view: "login" | "signup" | "dashboard";
  user: User | null;
  token: string | null;
  habits: HabitWithLogs[];
  isLoading: boolean;
  error: string | null;
  newHabitName: string;
};
