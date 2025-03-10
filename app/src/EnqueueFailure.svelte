<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { EnqueueMediaTicket } from "./proto/jungletv_pb";
    import { navigate } from "svelte-navigator";
    import ErrorMessage from "./ErrorMessage.svelte";
    import Wizard from "./Wizard.svelte";

    const dispatch = createEventDispatcher();

    export let ticket: EnqueueMediaTicket;

    function enqueueAnother() {
        dispatch("enqueueAnother");
    }

    function closeEnqueue() {
        navigate("/");
    }
</script>

<Wizard>
    <div slot="step-info">
        <h3 class="text-lg font-semibold leading-6 text-gray-900 dark:text-gray-200">Enqueue a video</h3>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
            <strong>Beware:</strong> if you just paid before the prices expired, it is possible your video was enqueued anyway.
            Double-check before trying again!
        </p>
    </div>
    <div slot="main-content">
        <div class="px-2 py-1 flex flex-row space-x-1 shadow-sm rounded-md border border-gray-300">
            <div class="w-32 flex-shrink-0">
                <img
                    alt="{ticket.getYoutubeVideoData().getTitle()} thumbnail"
                    src={ticket.getYoutubeVideoData().getThumbnailUrl()}
                />
            </div>
            <div class="flex flex-col flex-grow">
                <p>{ticket.getYoutubeVideoData().getTitle()}</p>
                <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{ticket.getYoutubeVideoData().getChannelTitle()}</p>
            </div>
        </div>
        <div class="mt-8">
            <ErrorMessage>
                Payment not received in time. If you did not make a payment yet, please try again.
            </ErrorMessage>
        </div>
    </div>
    <div slot="buttons">
        <button
            type="button"
            class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 hover:shadow ease-linear transition-all duration-150"
            on:click={closeEnqueue}
        >
            Cancel
        </button>
        <span class="mt-10 text-xs text-gray-400">Ticket ID: <span class="font-mono">{ticket.getId()}</span></span>
        <button
            type="submit"
            class="inline-flex float-right justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-yellow-600 hover:bg-yellow-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-yellow-500 hover:shadow ease-linear transition-all duration-150"
            on:click={enqueueAnother}
        >
            Try again
        </button>
    </div>
</Wizard>
