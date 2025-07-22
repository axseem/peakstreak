import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";
import { NavigateFx } from "../router";
import { Logout } from "../state";

// Helper for nav links
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
      type: "search",
      placeholder: "Search",
      class: "w-full bg-neutral-800 rounded-lg py-2.5 pl-11 pr-4 focus:outline-none placeholder-neutral-500"
    })
  ]);

const ProfileLink = (state: State): VNode<State> =>
  h("a",
    {
      href: `/@${state.user?.username}`,
      class: "flex items-center gap-3 group",
      onclick: (s: State, event: Event) => {
        event.preventDefault();
        if (s.user) {
          return [s, [NavigateFx, { path: `/@${s.user.username}` }]];
        }
        return s;
      }
    },
    [
      h("div", { class: "w-10 h-10 bg-neutral-700 rounded-full" }),
      h("span", { class: "text-neutral-300 group-hover:text-white" }, text("@" + state.user?.username || "Profile"))
    ]
  );

export const Sidebar = (state: State): VNode<State> =>
  h("aside", { class: "w-64 h-screen sticky top-0 bg-black flex-shrink-0 flex flex-col p-4 md:p-8 border-r border-neutral-900" }, [
    h("div", { class: "flex flex-col gap-4" }, [
      SearchBar(),
      h("nav", { class: "flex flex-col gap-4 mt-4" }, [
        NavLink({ path: `/@${state.user?.username}`, label: "My Profile", active: state.view === 'profile' && !!state.profileData?.isOwner }),
        NavLink({ path: "/leaderboard", label: "Leaderboard", active: state.view === 'leaderboard' }),
        NavLink({ path: "/explore", label: "Explore", active: state.view === 'explore' }),
      ])
    ]),
    h("div", { class: "mt-auto flex flex-col gap-4" }, [
      ProfileLink(state),
      h("button", {
        class: "text-left text-neutral-500 hover:text-red-400 transition-colors duration-200 pl-1",
        onclick: Logout
      }, text("Logout"))
    ])
  ]);
