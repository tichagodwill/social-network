<script lang="ts">
  import placeholder from '$lib/assets/angy.png'
  import { getLastDate } from '$lib/dateFormater'

  export let title = 'Big Man'
  export let subTitle = 'I want to be spiderman'
  export let image = ''
  export let imageAlt = ''
  export let date = new Date('2024-10-05T10:20:30Z')
  export let isSelected = false

  if (image.length == 0)
    image = placeholder

  const formmatedDate = getLastDate(date)

  $: bg = isSelected ? 'bg-blue-300' : 'hover:bg-gray-200 dark:hover:bg-gray-400'
</script>


<button class="flex flex-row items-center gap-2 mx-2 p-2 rounded-lg {bg}">
  <div>
    <img class="w-14 h-14 rounded-full" src={image} alt="{imageAlt} avatar" />
  </div>

  <div class="flex flex-col items-start gap-1">
    <div class="font-medium text-black">{title}</div>
    <div>{subTitle}</div>
  </div>

  <div class="ml-auto mb-auto pt-1 pr-1">{formmatedDate}</div>
</button>