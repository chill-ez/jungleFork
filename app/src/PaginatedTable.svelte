<script lang="ts">
  import { onMount } from "svelte";
  import { PaginationParameters } from "./proto/jungletv_pb";

  export let per_page = 10;
  export let cur_page = 0;
  export let title = "Items";
  export let loading_message = "Loading...";
  export let error_message = "Error loading results";
  export let no_items_message = "No items";
  export let no_results_message = "No results";
  export let search_query = "";
  export let column_count = 4;
  export let is_card = true;
  export let data_promise_factory: (pagParams: PaginationParameters) => Promise<[any[], number]>;
  let num_pages = 0;
  let last_page_items = 0;
  let waiting_for_page = -1;
  let dataPromise = new Promise<[any[], number]>((resolve, reject) => {});

  let paginationParams = new PaginationParameters();

  onMount(async () => {
    await changePage(0);
  });

  let updateDataCallback = async () => {
    await changePage(cur_page);
    if (last_page_items == 0 && cur_page > 0) {
      await changePage(cur_page - 1);
    }
  };

  // helpers for pagination control
  let page_nums = Array.from(new Array(num_pages).keys());
  let pages_to_show = {};
  let ellipsis_position = {};

  async function changePage(page: number) {
    waiting_for_page = page;
    paginationParams.setLimit(per_page);
    paginationParams.setOffset(page * per_page);
    dataPromise = data_promise_factory(paginationParams);
    let dataResponse = await dataPromise;
    if (waiting_for_page != page) {
      // we received data for some other page first
      return;
    }
    num_pages = Math.ceil(dataResponse[1] / per_page);
    cur_page = page;
    last_page_items = dataResponse[0].length;

    page_nums = Array.from(new Array(num_pages).keys());
    let ellipsis_shown = false;
    let pages_to_show_new = {};
    let ellipsis_position_new = {};
    let budget = 7;
    page_nums.forEach((page) => {
      if (page < 1 || page >= num_pages - Math.max(budget, 1) || (page > cur_page - 3 && page < cur_page + 3)) {
        budget--;
        pages_to_show_new[page] = true;
        ellipsis_shown = false;
      } else if (!ellipsis_shown) {
        ellipsis_position_new[page] = true;
        ellipsis_shown = true;
      }
    });
    pages_to_show = pages_to_show_new;
    ellipsis_position = ellipsis_position_new;
  }

  async function changePageWithElem(button: HTMLButtonElement, page: number) {
    button.classList.add("animate-pulse");
    button.style["animationDuration"] = "0.5s";
    await changePage(page);
    button.classList.remove("animate-pulse");
  }

  $: {
    search_query = search_query;
    if (cur_page < 0) {
      cur_page = 0;
    }
    changePage(cur_page);
  }
</script>

<div
  class="relative flex flex-col min-w-0 break-words w-full mb-6 {is_card
    ? 'shadow-lg rounded'
    : ''} bg-white dark:bg-gray-800"
>
  <div class="rounded-t mb-0 px-4 py-3 border-0">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full px-2 max-w-full flex-grow flex-1">
        <h3 class="font-semibold text-lg text-gray-800 dark:text-white">
          {title}
          {#if search_query}
            (searching for "{search_query}")
          {/if}
        </h3>
      </div>
    </div>
  </div>
  <div class="block w-full overflow-x-auto">
    <table class="items-center w-full bg-transparent border-collapse">
      <thead>
        <slot name="thead" />
      </thead>
      <tbody>
        {#await dataPromise}
          {#each Array.from(new Array(Math.max(last_page_items, 1)).keys()) as rownum}
            <tr>
              <td
                class="border-t-0 px-6 align-middle text-center border-l-0 border-r-0 text-s whitespace-no-wrap p-4"
                colspan={column_count}
              >
                {#if rownum === 0}
                  {loading_message}
                {:else}&nbsp;
                {/if}
              </td>
            </tr>
          {/each}
        {:then response}
          {#each response[0] as item}
            <slot name="item" {item} {updateDataCallback} />
          {:else}
            <td
              class="border-t-0 px-6 align-middle text-center border-l-0 border-r-0 text-s whitespace-no-wrap p-4"
              colspan={column_count}>{search_query != "" ? no_results_message : no_items_message}</td
            >
          {/each}
        {:catch}
          <tr>
            <td
              class="border-t-0 px-6 align-middle text-center border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-red-600"
              colspan={column_count}>{error_message}</td
            >
          </tr>
        {/await}
      </tbody>
    </table>
    {#if num_pages > 1}
      <div class="py-2 align-middle">
        <nav class="block">
          <ul class="flex px-6 rounded list-none flex-wrap">
            {#each page_nums as page}
              {#if pages_to_show[page]}
                <li>
                  <button
                    on:click={(event) => changePageWithElem(event.currentTarget, page)}
                    class="first:ml-0 text-xs font-semibold flex w-8 h-8 mx-1 p-0 rounded-full items-center justify-center leading-tight relative border border-solid border-profilerblue
                    {page === cur_page ? 'bg-blue-600 text-white' : ' bg-white text-blue-600'}
                  ">{page + 1}</button
                  >
                </li>
              {:else if ellipsis_position[page]}
                <li>
                  <i class="w-4 text-center fas fa-ellipsis-h align-bottom text-sm text-gray-400" />
                </li>
              {/if}
            {/each}
          </ul>
        </nav>
      </div>
    {/if}
  </div>
</div>
