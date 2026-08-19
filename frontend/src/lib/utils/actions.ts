import tippy from "tippy.js";
import 'tippy.js/animations/shift-away.css';
import type { Placement } from "tippy.js";
import type { Attachment } from "svelte/attachments";

export const attachTooltip: (tooltipOptions: { text: string, placement: Placement }) => Attachment<HTMLAnchorElement> = function (options) {
    return function (element) {
        tippy(element, { content: options.text, placement: options.placement, theme: 'tomato', animation: 'shift-away', offset: [10, 20] })
    }
}