import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";
import {
  HandleFormInput,
  HandleColorInput,
  CreateHabitFx,
  ShowAddHabitForm,
  HideAddHabitForm,
  ToggleEditMode,
  FollowUserFx,
  UnfollowUserFx,
  OpenFollowerList,
  UploadAvatarFx
} from "../state";
import { Input } from "../components/Input";
import { Button } from "../components/Button";
import { UserListPopup } from "../components/UserListPopup";
import { Avatar } from "../components/Avatar";
import { HabitCard } from "../components/HabitCard";



const InlineCreateHabitForm = (state: State): VNode<State> => h("form", {
  class: "flex flex-col items-center justify-center gap-4 p-8 w-full h-full",
  onsubmit: (s: State, event: Event) => {
    event.preventDefault();
    if (s.newHabitName.trim() && s.token) {
      return [s, [CreateHabitFx, { name: s.newHabitName, colorHue: s.newHabitColorHue, token: s.token }]];
    }
    return s;
  },
}, [
  Input({
    label: "What habit do you want to build?",
    id: "newHabitInline",
    placeholder: "e.g. Read for 15 minutes",
    value: state.newHabitName,
    oninput: HandleFormInput,
    disabled: state.isLoading,
    oncreate: (el: HTMLInputElement) => el.focus(),
    class: "text-center"
  }),
  Input({
    label: "Color Hue (0-360)",
    id: "newHabitColor",
    type: "number",
    min: "0",
    max: "360",
    value: state.newHabitColorHue,
    oninput: HandleColorInput,
    disabled: state.isLoading,
    class: "w-full text-center"
  }),
  h("div", { class: "flex gap-2 w-full mt-4" }, [
    Button({
      type: "button",
      onclick: HideAddHabitForm,
      class: "w-full bg-neutral-700 enabled:hover:bg-neutral-800 text-neutral-300"
    }, text("Cancel")),
    Button({
      type: "submit",
      disabled: state.isLoading || !state.newHabitName.trim(),
      class: "w-full"
    }, text("Add Habit")),
  ])
]);

const AddHabitCard = (state: State): VNode<State> => {
  if (state.isAddingHabit) {
    return h("div", {
      class: "bg-neutral-900 rounded-4xl flex items-center justify-center w-full",
      key: "add-habit-form"
    }, [
      InlineCreateHabitForm(state)
    ]);
  }

  return h("div", {
    key: "add-habit-button",
    class: "bg-transparent border-2 border-dashed border-neutral-800 rounded-4xl flex items-center justify-center min-h-[16rem] hover:bg-neutral-900 hover:border-neutral-700 transition-colors cursor-pointer",
    onclick: ShowAddHabitForm
  }, [
    h("svg", {
      xmlns: "http://www.w3.org/2000/svg",
      fill: "none",
      viewBox: "0 0 24 24",
      "stroke-width": 1.5,
      stroke: "currentColor",
      class: "w-12 h-12 text-neutral-700"
    }, [
      h("path", {
        "stroke-linecap": "round",
        "stroke-linejoin": "round",
        d: "M12 4.5v15m7.5-7.5h-15"
      })
    ])
  ]);
};

export const ProfileView = (state: State): VNode<State> => {
  if (state.isLoading && !state.profileData) {
    return h("p", {}, text("Loading profile..."));
  }

  if (!state.profileData) {
    return h("p", { class: "text-red-400" }, text(state.error || "Profile could not be loaded."));
  }

  const { user, habits, isOwner, followersCount, followingCount, isFollowing } = state.profileData;

  const FollowButton = () => Button({
    onclick: (s: State) => {
      if (s.isLoading || !s.token || !s.profileData) return s;
      const fx = s.profileData.isFollowing ? UnfollowUserFx : FollowUserFx;
      return [s, [fx, { username: s.profileData.user.username, token: s.token }]];
    },
    class: isFollowing ? "bg-neutral-700 enabled:hover:bg-neutral-800 text-neutral-300" : "",
    disabled: state.isLoading,
  }, text(isFollowing ? "Following" : "Follow"));

  const EditButton = () => Button({
    onclick: ToggleEditMode,
    class: "self-start"
  }, text(state.isEditing ? "Done" : "Edit"));


  return h("div", { class: "flex flex-col gap-8 w-full" }, [
    isOwner && h("input", {
      id: "avatar-upload",
      type: "file",
      accept: "image/png, image/jpeg",
      class: "hidden",
      onchange: (s: State, event: Event) => {
        const file = (event.target as HTMLInputElement).files?.[0];
        if (file && s.token) {
          return [s, [UploadAvatarFx, { file, token: s.token }]];
        }
        return s;
      }
    }),
    h("header", { class: "flex justify-between items-start" }, [
      h("div", { class: "flex items-center gap-6" }, [
        h("div", { class: "relative" }, [
          Avatar({ src: user.avatarUrl, username: user.username, sizeClass: "w-20 h-20" }),
          isOwner && h("label", {
            for: "avatar-upload",
            class: "absolute -bottom-1 -right-1 bg-neutral-800 p-2 rounded-full cursor-pointer hover:bg-neutral-700 border-2 border-black transition-colors",
            title: "Change profile picture"
          }, h("svg", { class: "w-4 h-4", xmlns: "http://www.w3.org/2000/svg", fill: "none", viewBox: "0 0 24 24", "stroke-width": "1.5", stroke: "currentColor" },
            h("path", { "stroke-linecap": "round", "stroke-linejoin": "round", d: "M6.827 6.175A2.31 2.31 0 015.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 00-1.134-.175 2.31 2.31 0 01-1.64-1.055l-.822-1.316a2.192 2.192 0 00-1.736-1.039 48.776 48.776 0 00-5.232 0 2.192 2.192 0 00-1.736 1.039l-.821 1.316zM12 15a3.75 3.75 0 100-7.5 3.75 3.75 0 000 7.5z" })
          )),
        ]),
        h("div", { class: "flex flex-col gap-2" }, [
          h("h1", { class: "text-3xl font-bold" }, text("@" + user.username)),
          h("div", { class: "flex gap-4 text-neutral-400" }, [
            h("button", {
              class: "hover:underline disabled:no-underline disabled:cursor-default",
              disabled: followersCount === 0,
              onclick: (s: State) => OpenFollowerList(s, { type: 'followers', username: user.username })
            }, text(`${followersCount} Followers`)),
            h("button", {
              class: "hover:underline disabled:no-underline disabled:cursor-default",
              disabled: followingCount === 0,
              onclick: (s: State) => OpenFollowerList(s, { type: 'following', username: user.username })
            }, text(`${followingCount} Following`)),
          ]),
        ]),
      ]),
      isOwner
        ? (habits.length > 0 && EditButton())
        : FollowButton()
    ]),

    h("div", { class: "flex flex-col gap-4" }, [
      h("div", { class: "flex justify-between items-center" }, [
        h("h2", { class: "text-2xl font-bold" }, text("Habits")),
      ]),
      h("div", { class: "flex flex-col gap-8" }, [
        ...habits.map(habit => HabitCard(habit, isOwner, state.token, state.isEditing)),

        isOwner && !state.isEditing ? AddHabitCard(state) : null,

        habits.length === 0 && !isOwner
          ? h("p", { class: "text-neutral-400" }, text("This user hasn't added any habits yet."))
          : null
      ])
    ]),
    UserListPopup({ followerList: state.followerList })
  ]);
};
