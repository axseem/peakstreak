import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";
import {
  HandleFormInput,
  HandleColorInput,
  HandleIsBooleanInput,
  CreateHabitFx,
  ShowAddHabitForm,
  HideAddHabitForm,
  FollowUserFx,
  UnfollowUserFx,
  OpenFollowerList,
  UploadAvatarFx,
} from "../state";
import { NavigateFx } from "../router.ts";
import { Input } from "../components/Input";
import { Button } from "../components/Button";
import { UserListPopup } from "../components/UserListPopup";
import { Avatar } from "../components/Avatar";
import { HabitCard } from "../components/HabitCard";

const InlineCreateHabitForm = (state: State): VNode<State> =>
  h(
    "form",
    {
      class:
        "flex flex-col items-center justify-center gap-4 p-8 w-full h-full",
      onsubmit: (s: State, event: Event) => {
        event.preventDefault();
        if (s.newHabitName.trim() && s.token) {
          return [
            s,
            [
              CreateHabitFx,
              {
                name: s.newHabitName,
                colorHue: s.newHabitColorHue,
                isBoolean: s.newHabitIsBoolean,
                token: s.token,
              },
            ],
          ];
        }
        return s;
      },
    },
    [
      Input({
        label: "What habit do you want to build?",
        id: "newHabitInline",
        placeholder: "e.g. Read for 15 minutes",
        value: state.newHabitName,
        oninput: HandleFormInput,
        disabled: state.isLoading,
        oncreate: (el: HTMLInputElement) => el.focus(),
        class: "text-center",
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
        class: "w-full text-center",
      }),
      h(
        "div",
        { class: "flex items-center self-start gap-2 text-neutral-400 mt-2" },
        [
          h("input", {
            type: "checkbox",
            id: "newHabitIsBoolean",
            class:
              "w-4 h-4 rounded bg-neutral-700 border-neutral-600 text-blue-500 focus:ring-blue-600",
            checked: state.newHabitIsBoolean,
            oninput: HandleIsBooleanInput,
          }),
          h(
            "label",
            { for: "newHabitIsBoolean" },
            text("Simple (Yes/No) habit"),
          ),
        ],
      ),
      h("div", { class: "flex gap-2 w-full mt-4" }, [
        Button(
          {
            type: "button",
            onclick: HideAddHabitForm,
            class:
              "w-full bg-neutral-700 enabled:hover:bg-neutral-800 text-neutral-300",
          },
          text("Cancel"),
        ),
        Button(
          {
            type: "submit",
            disabled: state.isLoading || !state.newHabitName.trim(),
            class: "w-full",
          },
          text("Add Habit"),
        ),
      ]),
    ],
  );

const AddHabitCard = (state: State): VNode<State> => {
  if (state.isAddingHabit) {
    return h(
      "div",
      {
        class:
          "bg-neutral-900 rounded-4xl flex items-center justify-center w-full",
        key: "add-habit-form",
      },
      [InlineCreateHabitForm(state)],
    );
  }

  return h(
    "div",
    {
      key: "add-habit-button",
      class:
        "bg-transparent border-2 border-dashed border-neutral-800 rounded-4xl flex items-center justify-center min-h-[16rem] hover:bg-neutral-900 hover:border-neutral-700 transition-colors cursor-pointer",
      onclick: ShowAddHabitForm,
    },
    [
      h("iconify-icon", {
        icon: "material-symbols:add",
        class: "text-neutral-700",
        width: 48,
        height: 48,
      }),
    ],
  );
};

export const ProfileView = (state: State): VNode<State> => {
  if (state.isLoading && !state.profileData) {
    return h("p", {}, text("Loading profile..."));
  }

  if (!state.profileData) {
    return h(
      "p",
      { class: "text-red-400" },
      text(state.error || "Profile could not be loaded."),
    );
  }

  const { user, habits, isOwner, followersCount, followingCount, isFollowing } =
    state.profileData;

  const FollowButton = () =>
    Button(
      {
        onclick: (s: State) => {
          if (s.isLoading || !s.token || !s.profileData) return s;
          const fx = s.profileData.isFollowing ? UnfollowUserFx : FollowUserFx;
          return [
            s,
            [fx, { username: s.profileData.user.username, token: s.token }],
          ];
        },
        class:
          (isFollowing
            ? "bg-neutral-700 enabled:hover:bg-neutral-800 text-neutral-300"
            : "") + " flex-grow md:flex-grow-0",
        disabled: state.isLoading,
      },
      text(isFollowing ? "Following" : "Follow"),
    );

  const SettingsButton = () =>
    Button(
      {
        onclick: (s: State) => [s, [NavigateFx, { path: "/settings" }]],
        class:
          "bg-neutral-800 enabled:hover:bg-neutral-700 text-neutral-300 px-3 py-3 flex",
        title: "Account Settings",
      },
      h("iconify-icon", {
        icon: "material-symbols:settings-outline",
        width: 24,
        height: 24,
      }),
    );

  return h("div", { class: "flex flex-col gap-8 w-full" }, [
    isOwner &&
      h("input", {
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
        },
      }),
    h("header", { class: "flex flex-row justify-between items-center gap-4" }, [
      h("div", { class: "flex items-center gap-4" }, [
        h("div", { class: "relative flex-shrink-0" }, [
          Avatar({
            src: user.avatarUrl,
            username: user.username,
            sizeClass: "w-16 h-16 md:w-20 md:h-20",
          }),
          isOwner &&
            h(
              "label",
              {
                for: "avatar-upload",
                class:
                  "flex absolute -bottom-1 -right-1 bg-neutral-800 p-2 rounded-full cursor-pointer hover:bg-neutral-700 border-2 border-black transition-colors",
                title: "Change profile picture",
              },
              h("iconify-icon", {
                icon: "material-symbols:photo-camera-outline",
                width: 16,
                height: 16,
              }),
            ),
        ]),
        h("div", { class: "flex flex-col gap-1" }, [
          h(
            "h1",
            { class: "text-2xl md:text-3xl font-bold" },
            text("@" + user.username),
          ),
          h(
            "div",
            { class: "flex flex-wrap gap-x-4 gap-y-1 text-neutral-400" },
            [
              h(
                "button",
                {
                  class:
                    "hover:underline disabled:no-underline disabled:cursor-default",
                  disabled: followersCount === 0,
                  onclick: (s: State) =>
                    OpenFollowerList(s, {
                      type: "followers",
                      username: user.username,
                    }),
                },
                text(`${followersCount} Followers`),
              ),
              h(
                "button",
                {
                  class:
                    "hover:underline disabled:no-underline disabled:cursor-default",
                  disabled: followingCount === 0,
                  onclick: (s: State) =>
                    OpenFollowerList(s, {
                      type: "following",
                      username: user.username,
                    }),
                },
                text(`${followingCount} Following`),
              ),
            ],
          ),
        ]),
      ]),
      h("div", { class: "flex gap-2" }, [
        !isOwner && FollowButton(),
        isOwner && SettingsButton(),
      ]),
    ]),

    h("div", { class: "flex flex-col gap-4" }, [
      h("div", { class: "flex justify-between items-center" }, [
        h("h2", { class: "text-2xl font-bold" }, text("Habits")),
      ]),
      h("div", { class: "flex flex-col gap-8" }, [
        ...habits.map((habit) =>
          HabitCard({
            habit,
            isOwner,
            token: state.token,
            isEditing: state.editingHabit?.id === habit.id,
            activeHabitMenuId: state.activeHabitMenuId,
            editingHabit: state.editingHabit,
          }),
        ),

        isOwner && !state.editingHabit ? AddHabitCard(state) : null,

        habits.length === 0 && !isOwner
          ? h(
              "p",
              { class: "text-neutral-400" },
              text("This user hasn't added any habits yet."),
            )
          : null,
      ]),
    ]),
    UserListPopup({ followerList: state.followerList }),
  ]);
};
