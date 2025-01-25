<script lang="ts">
    import { onMount } from 'svelte';
    import { Card } from 'flowbite-svelte';
    import { fade } from 'svelte/transition';

    export let events: any[] = [];
    export let selectedDate: Date = new Date();
    export let onDateSelect: (date: Date) => void;

    let currentMonth = new Date();
    let weeks: Date[][] = [];
    let monthEvents: { [key: string]: any[] } = {};

    $: {
        generateCalendar(currentMonth);
        groupEventsByDate();
    }

    function generateCalendar(date: Date) {
        const year = date.getFullYear();
        const month = date.getMonth();
        const firstDay = new Date(year, month, 1);
        const lastDay = new Date(year, month + 1, 0);
        
        let currentDate = new Date(firstDay);
        currentDate.setDate(currentDate.getDate() - firstDay.getDay());
        
        weeks = [];
        let currentWeek: Date[] = [];
        
        while (currentDate <= lastDay || currentWeek.length > 0) {
            if (currentWeek.length === 7) {
                weeks.push(currentWeek);
                currentWeek = [];
            }
            
            currentWeek.push(new Date(currentDate));
            currentDate.setDate(currentDate.getDate() + 1);
            
            if (currentDate > lastDay && currentWeek.length < 7) {
                while (currentWeek.length < 7) {
                    currentWeek.push(new Date(currentDate));
                    currentDate.setDate(currentDate.getDate() + 1);
                }
                weeks.push(currentWeek);
                break;
            }
        }
    }

    function groupEventsByDate() {
        monthEvents = {};
        events.forEach(event => {
            const date = new Date(event.eventDate);
            const dateKey = date.toISOString().split('T')[0];
            if (!monthEvents[dateKey]) {
                monthEvents[dateKey] = [];
            }
            monthEvents[dateKey].push(event);
        });
    }

    function previousMonth() {
        currentMonth.setMonth(currentMonth.getMonth() - 1);
        currentMonth = new Date(currentMonth);
    }

    function nextMonth() {
        currentMonth.setMonth(currentMonth.getMonth() + 1);
        currentMonth = new Date(currentMonth);
    }

    function isToday(date: Date): boolean {
        const today = new Date();
        return date.toDateString() === today.toDateString();
    }

    function isSelected(date: Date): boolean {
        return date.toDateString() === selectedDate.toDateString();
    }

    function isCurrentMonth(date: Date): boolean {
        return date.getMonth() === currentMonth.getMonth();
    }

    function hasEvents(date: Date): boolean {
        const dateKey = date.toISOString().split('T')[0];
        return !!monthEvents[dateKey]?.length;
    }

    function getEventCount(date: Date): number {
        const dateKey = date.toISOString().split('T')[0];
        return monthEvents[dateKey]?.length || 0;
    }

    function handleDateClick(date: Date) {
        selectedDate = date;
        onDateSelect(date);
    }

    const monthNames = [
        "January", "February", "March", "April", "May", "June",
        "July", "August", "September", "October", "November", "December"
    ];
</script>

<style>
    .calendar {
        @apply select-none;
    }

    .calendar-header {
        @apply flex items-center justify-between mb-4;
    }

    .calendar-grid {
        @apply grid grid-cols-7 gap-1;
    }

    .calendar-day {
        @apply relative flex flex-col items-center justify-center p-2 rounded-lg transition-all duration-200 cursor-pointer;
    }

    .calendar-day:hover:not(.outside-month) {
        @apply bg-blue-50 dark:bg-blue-900;
    }

    .day-number {
        @apply text-sm font-medium;
    }

    .outside-month {
        @apply text-gray-400 dark:text-gray-600;
    }

    .today {
        @apply bg-blue-100 dark:bg-blue-800;
    }

    .selected {
        @apply bg-blue-500 text-white dark:bg-blue-600;
    }

    .has-events::after {
        content: '';
        @apply absolute bottom-1 w-1.5 h-1.5 rounded-full bg-green-500;
    }

    .selected.has-events::after {
        @apply bg-white;
    }

    .weekday {
        @apply text-center text-sm font-medium text-gray-500 dark:text-gray-400 py-2;
    }

    .month-nav {
        @apply p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors duration-200;
    }
</style>

<Card class="calendar">
    <div class="calendar-header">
        <button class="month-nav" on:click={previousMonth}>
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
            </svg>
        </button>
        <h3 class="text-lg font-semibold">
            {monthNames[currentMonth.getMonth()]} {currentMonth.getFullYear()}
        </h3>
        <button class="month-nav" on:click={nextMonth}>
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
            </svg>
        </button>
    </div>

    <div class="calendar-grid">
        {#each ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'] as day}
            <div class="weekday">{day}</div>
        {/each}
        
        {#each weeks as week}
            {#each week as date}
                <div
                    class="calendar-day {!isCurrentMonth(date) ? 'outside-month' : ''} 
                           {isToday(date) ? 'today' : ''} 
                           {isSelected(date) ? 'selected' : ''} 
                           {hasEvents(date) ? 'has-events' : ''}"
                    on:click={() => handleDateClick(date)}
                >
                    <span class="day-number">{date.getDate()}</span>
                    {#if hasEvents(date)}
                        <div 
                            class="text-xs {isSelected(date) ? 'text-white' : 'text-gray-600 dark:text-gray-400'}"
                            transition:fade
                        >
                            {getEventCount(date)} {getEventCount(date) === 1 ? 'event' : 'events'}
                        </div>
                    {/if}
                </div>
            {/each}
        {/each}
    </div>
</Card> 