import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";
import { NavigateFx } from "../router";
import { Logout, ToggleProfileMenu, CloseProfileMenu } from "../state";
import { Popup } from "./Popup";
import { Avatar } from "./Avatar";

const NavLink = ({ path, label, active }: { path: string, label: string, active: boolean }): VNode<State> =>
  h("a",
    {
      href: path,
      class: `text-lg ${active ? 'text-white' : 'text-neutral-400'} hover:text-white transition-colors duration-200`,
      onclick: (state: State, event: Event) => {
        event.preventDefault();
        return [state, [NavigateFx, { path }]];
      }
    },
    text(label)
  );

const SearchBar = (): VNode<State> =>
  h("div", { class: "relative" }, [
    h("span", { class: "absolute inset-y-0 left-0 flex items-center pl-4" },
      h("svg", { class: "h-5 w-5 text-neutral-500", viewBox: "0 0 20 20", fill: "currentColor" },
        h("path", { "fill-rule": "evenodd", d: "M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z", "clip-rule": "evenodd" })
      )
    ),
    h("input", {
      placeholder: "Search users...",
      class: "w-full bg-neutral-800 rounded-lg py-2.5 pl-11 pr-4 focus:outline-none placeholder-neutral-500",
      onkeydown: (state: State, event: KeyboardEvent) => {
        if (event.key === 'Enter') {
          const query = (event.target as HTMLInputElement).value.trim();
          if (query) {
            (event.target as HTMLInputElement).blur();
            return [state, [NavigateFx, { path: `/search?q=${encodeURIComponent(query)}` }]];
          }
        }
        return state;
      }
    })
  ]);

const ProfileMenu = (state: State): VNode<State> => {
  const LogoutButton = (): VNode<State> =>
    h("button", {
      class: "w-full text-left px-4 py-2 text-sm text-red-400 hover:bg-neutral-800 flex items-center gap-2 transition-colors",
      onclick: Logout
    }, [
      h("svg", { class: "w-4 h-4", xmlns: "http://www.w3.org/2000/svg", fill: "none", viewBox: "0 0 24 24", "stroke-width": "1.5", stroke: "currentColor" },
        h("path", { "stroke-linecap": "round", "stroke-linejoin": "round", d: "M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75" })
      ),
      text("Logout")
    ]);

  return h("div", { class: "relative" }, [
    h("div", { class: "flex items-center justify-between p-2 rounded-lg group hover:bg-neutral-800 transition-colors" }, [
      h("a", {
        href: `/@${state.user?.username}`,
        class: "flex items-center gap-3 flex-grow min-w-0",
        onclick: (s: State, event: Event) => {
          event.preventDefault();
          if (s.user) {
            return [CloseProfileMenu(s), [NavigateFx, { path: `/@${s.user.username}` }]];
          }
          return s;
        }
      }, [
        Avatar({ src: state.user?.avatarUrl, username: state.user?.username, sizeClass: "w-10 h-10" }),
        h("span", { class: "text-neutral-300 truncate" }, text("@" + state.user?.username || "Profile"))
      ]),
      h("button", {
        class: "text-neutral-400 hover:text-white p-1 rounded-md flex-shrink-0",
        onclick: ToggleProfileMenu,
      }, h("svg", { class: "w-6 h-6", fill: "none", viewBox: "0 0 24 24", "stroke-width": "1.5", stroke: "currentColor" },
        h("path", { "stroke-linecap": "round", "stroke-linejoin": "round", d: "M6.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0zM13.5 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0zM20.25 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0z" })
      ))
    ]),

    Popup({
      isOpen: state.isProfileMenuOpen,
      onClose: CloseProfileMenu,
      class: "bottom-full w-full mb-2 py-1"
    }, [
      LogoutButton()
    ])
  ]);
};

export const Sidebar = (state: State): VNode<State> =>
  h("aside", { class: "w-64 h-screen sticky top-0 bg-black flex-shrink-0 flex flex-col p-4 md:p-8 border-r border-neutral-900" }, [
    h("div", { class: "flex flex-col gap-4 flex-grow" }, [
      SearchBar(),
      h("nav", { class: "flex flex-col gap-4 mt-4" }, [
        NavLink({ path: "/leaderboard", label: "Leaderboard", active: state.view === 'leaderboard' }),
        NavLink({ path: "/explore", label: "Explore", active: state.view === 'explore' }),
      ])
    ]),
    h("div", { class: "mt-auto" }, [
      ProfileMenu(state),
    ])
  ]);
