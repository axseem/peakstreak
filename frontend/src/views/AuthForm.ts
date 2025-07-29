import { h, text, type VNode, type Dispatchable } from "hyperapp";
import type { State } from "../types";
import { LoginFx, SignUpFx } from "../state";
import { NavigateFx } from "../router";
import { Input } from "../components/Input";
import { Button } from "../components/Button";

export const AuthForm =
  (props: { isLogin: boolean }) =>
  (state: State): VNode<State> => {
    const buttonText = state.isLoading
      ? "Loading..."
      : props.isLogin
        ? "Log In"
        : "Create Account";
    const questionText = props.isLogin
      ? "Don't have an account? "
      : "Already have an account? ";
    const linkText = props.isLogin ? "Sign Up" : "Log In";
    const pathToggle = props.isLogin ? "/signup" : "/login";

    const onsubmit = (
      state: State,
      event: SubmitEvent,
    ): Dispatchable<State, any> => {
      event.preventDefault();
      const formData = new FormData(event.target as HTMLFormElement);
      const data = Object.fromEntries(formData.entries());
      if (props.isLogin) {
        const loginPayload = {
          identifier: data.username,
          password: data.password,
        };
        return [state, [LoginFx, loginPayload]];
      }
      return [state, [SignUpFx, data]];
    };

    const footer = h<State>("div", { class: "flex gap-2" }, [
      h("span", { class: "text-neutral-400" }, text(questionText)),
      h(
        "a",
        {
          class: "underline cursor-pointer",
          onclick: (state: State) => [
            state,
            [NavigateFx, { path: pathToggle }],
          ],
          href: pathToggle,
        },
        text(linkText),
      ),
    ]);

    return h(
      "div",
      {
        class:
          "w-full h-screen flex flex-col items-center justify-center gap-4 p-4",
      },
      [
        h("form", { class: "flex flex-col gap-4 w-full max-w-sm", onsubmit }, [
          Input({
            label: "Username",
            id: "username",
            type: "text",
            required: true,
          }),
          !props.isLogin &&
            Input({
              label: "Email",
              id: "email",
              type: "email",
              required: true,
            }),
          Input({
            label: "Password",
            id: "password",
            type: "password",
            required: true,
          }),
          state.error
            ? h<State>("p", { class: "text-red-400" }, text(state.error))
            : null,
          Button(
            { type: "submit", disabled: state.isLoading },
            text(buttonText),
          ),
        ]),
        footer,
      ],
    );
  };
