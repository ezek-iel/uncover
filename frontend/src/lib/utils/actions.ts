import tippy from "tippy.js";
import 'tippy.js/animations/shift-away.css';
import type { Placement } from "tippy.js";
import type { Attachment } from "svelte/attachments";
import type { AIResponse } from "./fetch";
import MarkdownIt from "markdown-it";

export const attachTooltip: (tooltipOptions: { text: string, placement: Placement }) => Attachment<HTMLAnchorElement> = function (options) {
    return function (element) {
        const tp = tippy(element, { content: options.text, placement: options.placement, theme: 'tomato', animation: 'shift-away', offset: [10, 20] })
        return () => { tp.destroy() }
    }
}

const markdownIt = new MarkdownIt('commonmark', { breaks: true })

// Only allow safe link protocols (the default already blocks `javascript:` etc.,
// but be explicit: http/https/mailto, anchors, and relative paths).
markdownIt.validateLink = (url: string) => /^(?:(?:https?|mailto):|#|\/)/i.test(url)

export const styleAIResponse: (total: AIResponse) => Attachment<HTMLDivElement> = function (total) {
    return function (element) {

        // This action is re-invoked every time a new token streams in (the caller
        // re-runs it reactively), so always start from a clean slate — otherwise
        // every streamed token would duplicate all previously rendered content.
        element.replaceChildren()

        function createDiv(className: string, text: string): HTMLDivElement {
            const div = document.createElement('div')
            div.classList.add(className)
            div.textContent = text

            return div
        }

        // Group consecutive tokens of the same type into a single container.
        let previousType = ""
        let previousContainer: HTMLDivElement | null = null

        for (const token of total) {
            if (token.type === "tool-response") continue;

            if (previousType !== token.type) {
                previousContainer = createDiv(token.type, token.message)
                previousType = token.type
                element.appendChild(previousContainer)
            } else if (previousContainer) {
                // Plain concatenation: the backend preserves newlines inside a
                // chunk via SSE multi-line `data:` fields, so consecutive
                // tokens are already exact fragments of the final text.
                previousContainer.textContent += token.message
            }
        }

        // Render tool-request queries (tab-separated, possibly with a trailing tab)
        // as a bulleted list.
        for (const container of element.querySelectorAll("div.tool-request")) {
            const queries = (container.textContent ?? "").split("\t").filter(q => q.length > 0)

            container.textContent = ""
            const ul = document.createElement('ul')

            for (const query of queries) {
                const li = document.createElement('li')
                li.textContent = query
                ul.appendChild(li)
            }

            container.appendChild(ul)
        }

        // Render markdown for content blocks. Use textContent (not innerHTML) so the
        // raw markdown isn't HTML-escaped before being parsed.
        for (const container of element.querySelectorAll("div.content")) {
            container.innerHTML = markdownIt.render(container.textContent ?? "")
        }
    }
}
