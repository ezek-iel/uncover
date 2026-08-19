<script lang="ts">
    import logoIconOnly from "../../assets/logoIconOnly.svg";
    import IconHome from "~icons/solar/home-2-outline";
    import IconChatDots from "~icons/solar/chat-dots-outline";
    import IconSettingsDuotone from "~icons/solar/settings-line-duotone";
    import { attachTooltip } from "../utils/actions";
</script>

<aside class="sidebar">
    <div class="sidebar-top">
        <div>
            <img src={logoIconOnly} alt="" class="logo" />
        </div>
        <div>
            <ul class="sidebar-links">
                <li>
                    <a
                        href="/"
                        {@attach attachTooltip({
                            text: "Create new chat",
                            placement: "right",
                        })}
                    >
                        <IconHome class="text-slate-700 size-8" />
                    </a>
                </li>
                <li>
                    <a href="/">
                        <IconChatDots
                            class="text-slate-700 size-8"
                            {@attach attachTooltip({
                                text: "View Previous Chats",
                                placement: "right",
                            })}
                        />
                    </a>
                </li>
            </ul>
        </div>
    </div>

    <div class="sidebar-bottom">
        <ul class="sidebar-links">
            <li><IconSettingsDuotone class="text-slate-700 size-8" /></li>
            <li>
                <img
                    src="https://cdn.jsdelivr.net/gh/alohe/memojis/png/upstream_4.png"
                    alt=""
                    class="profile"
                />
            </li>
        </ul>
    </div>
</aside>

<style>
    @reference "../../app.css";

    .profile {
        @apply size-12 rounded-full shadow-inner;
    }

    .sidebar-top {
        @apply flex items-center flex-col gap-8;
    }

    .sidebar-links {
        @apply flex flex-col items-center gap-6;
    }

    .sidebar-links li {
        @apply rounded-lg p-2.5 hover:bg-slate-200 transition-colors;
    }

    .sidebar {
        @apply h-screen w-fit py-8 px-2 flex flex-col justify-between border-r border-r-slate-300;
    }

    .sidebar-top {
        @apply flex flex-col items-center;
    }
</style>
