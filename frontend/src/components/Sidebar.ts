import { h, text, type VNode } from "hyperapp";
import type { State } from "../types";
import { NavigateFx } from "../router";
import { Logout, ToggleProfileMenu, CloseProfileMenu } from "../state";
import { Popup } from "./Popup";
import { Avatar } from "./Avatar";
import { Menu, MenuItem } from "./Menu";
import { twMerge } from "tailwind-merge";

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
      h("iconify-icon", { icon: "material-symbols:search", class: "text-neutral-500", width: 20, height: 20 })
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
        class: "text-neutral-400 hover:text-white p-1 rounded-md flex-shrink-0 flex",
        onclick: ToggleProfileMenu,
      }, h("iconify-icon", { icon: "material-symbols:more-horiz", width: 24, height: 24 }))
    ]),

    Popup({
      isOpen: state.isProfileMenuOpen,
      onClose: CloseProfileMenu,
      class: "bottom-full w-full mb-2"
    }, Menu({}, [
      MenuItem({
        onclick: (s: State) => [CloseProfileMenu(s), [NavigateFx, { path: "/settings" }]],
      }, [
        h("iconify-icon", { icon: "material-symbols:settings-outline", width: 16, height: 16 }),
        text("Settings")
      ]),
      MenuItem({
        class: "text-red-400 hover:bg-red-500/10 hover:text-red-300",
        onclick: Logout
      }, [
        h("iconify-icon", { icon: "material-symbols:logout", width: 16, height: 16 }),
        text("Logout")
      ])
    ]))
  ]);
};

// --- Mobile Nav ---
const MobileNavItem = ({ path, label, icon, active }: { path: string, label: string, icon: VNode<State>, active: boolean }): VNode<State> =>
  h("a", {
    href: path,
    class: twMerge("flex flex-col items-center justify-center gap-1 w-full pt-2 pb-1 transition-colors", active ? "text-white" : "text-neutral-400 hover:text-white"),
    onclick: (state: State, event: Event) => {
      event.preventDefault();
      return [state, [NavigateFx, { path }]];
    }
  }, [
    icon,
    h("span", { class: "text-xs" }, text(label))
  ]);

const MobileNav = (state: State): VNode<State> | null => {
  if (!state.user) return null;

  return h("nav", {
    class: "md:hidden fixed bottom-0 left-0 right-0 bg-black border-t border-neutral-900 z-30"
  }, [
    h("div", { class: "flex justify-around items-center h-16" }, [
      MobileNavItem({
        path: "/leaderboard", label: "Leaders", active: state.view === 'leaderboard',
        icon: h("iconify-icon", { icon: "material-symbols:leaderboard-outline", width: 24, height: 24 })
      }),
      MobileNavItem({
        path: "/explore", label: "Explore", active: state.view === 'explore',
        icon: h("iconify-icon", { icon: "material-symbols:explore-outline", width: 24, height: 24 })
      }),
      MobileNavItem({
        path: `/search`, label: "Search", active: state.view === 'search',
        icon: h("iconify-icon", { icon: "material-symbols:search", width: 24, height: 24 })
      }),
      MobileNavItem({
        path: `/@${state.user.username}`, label: "Profile", active: state.view === 'profile' && state.profileData?.isOwner === true,
        icon: h("div", { width: 24, height: 24 }, Avatar({ src: state.user?.avatarUrl, username: state.user?.username, sizeClass: "w-6 h-6" }))
      })
    ])
  ]);
};

export const Sidebar = (state: State): VNode<State> =>
  h("div", {}, [
    h("aside", { class: "hidden md:flex w-64 h-screen sticky top-0 bg-black flex-shrink-0 flex-col p-8 border-r border-neutral-900" }, [
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
    ]),
    MobileNav(state)
  ]);
