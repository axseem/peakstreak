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

export type HabitWithStatus = Habit & { loggedToday: boolean };

export type State = {
  view: "login" | "signup" | "dashboard";
  user: User | null;
  token: string | null;
  habits: HabitWithStatus[];
  isLoading: boolean;
  error: string | null;
  newHabitName: string;
};
