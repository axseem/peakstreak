import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";
import { Button } from "../components/Button";
import { DeleteAccountFx, SetError } from "../state";

export const SettingsView = (state: State): VNode<State> => {
  const handleDelete = (s: State) => {
    if (
      s.token &&
      window.confirm(
        "Are you sure you want to permanently delete your account? This action cannot be undone.",
      )
    ) {
      return [s, [DeleteAccountFx, { token: s.token }]];
    }
    return s;
  };

  return h(
    "div",
    {
      class: "flex flex-col gap-8 w-full",
      onremove: (state: State) => SetError(state, null),
    },
    [
      h("h1", { class: "text-3xl font-bold" }, text("Settings")),

      h(
        "div",
        {
          class:
            "flex flex-col gap-4 p-4 md:p-6 rounded-2xl bg-neutral-900/50 border border-neutral-800/50",
        },
        [
          h("h2", { class: "text-xl font-bold" }, text("Account")),
          state.error
            ? h("p", { class: "text-red-400" }, text(state.error))
            : null,

          h(
            "div",
            {
              class:
                "pt-4 mt-4 border-t border-neutral-800/50 flex flex-col items-start gap-4",
            },
            [
              h(
                "h3",
                { class: "font-bold text-lg text-red-400" },
                text("Danger Zone"),
              ),
              h(
                "p",
                { class: "text-neutral-400" },
                text(
                  "Permanently delete your account and all associated data. This action is irreversible.",
                ),
              ),
              Button(
                {
                  class:
                    "bg-red-500/20 text-red-400 enabled:hover:bg-red-500/30",
                  onclick: handleDelete,
                },
                text("Delete My Account"),
              ),
            ],
          ),
        ],
      ),
    ],
  );
};
