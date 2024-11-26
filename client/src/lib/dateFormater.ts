const units: { label: string; seconds: number }[] = [
    { label: 'years', seconds: 31536000 },
    { label: 'months', seconds: 2592000 },
    { label: 'days', seconds: 86400 },
    { label: 'hours', seconds: 3600 },
    { label: 'minutes', seconds: 60 },
    { label: 'seconds', seconds: 1 }
];

export function getFormattedDate(inputDate: Date | string | undefined) {
    if (!inputDate) {
        return { formated: 'No date', diff: 'No date' };
    }

    // Convert string to Date if necessary
    const date = typeof inputDate === 'string' ? new Date(inputDate) : inputDate;
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
        return { formated: 'Invalid date', diff: 'Invalid date' };
    }

    const formattedDate = date.toDateString(); // Format: "Tue Sep 24 2024"

    // diff
    let suffix = 'ago';
    let now = new Date();

    if (now < date) {
        const tmp = now;
        suffix = 'ahead';
        now = date;
        inputDate = tmp;
    }

    const diffSeconds = Math.floor((now.getTime() - date.getTime()) / 1000) + 1;

    const unit = units.find(u => diffSeconds >= u.seconds);
    if (!unit) {
        return { formated: formattedDate, diff: 'just now' };
    }

    const value = Math.round(diffSeconds / unit.seconds);
    const timeAgo = `${value} ${unit.label} ${suffix}`;

    return { formated: formattedDate, diff: timeAgo };
}

export function getLastDate(input: Date | string | undefined) {
    if (!input) {
        return 'No date';
    }

    const date = typeof input === 'string' ? new Date(input) : input;
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
        return 'Invalid date';
    }

    const now = new Date();
    const diff = Math.floor((now.getTime() - date.getTime()) / 1000) + 1;

    if (diff < 3) {
        return 'now';
    }

    if (diff < 60) { // 32s
        return `${diff}s`;
    }

    if (diff < 3600) { // <1hour: 32min
        const minutes = Math.floor(diff / 60);
        return `${minutes}m`;
    }

    if (diff < 86400) { // <1day: 12:33pm
        return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: 'numeric', hour12: true });
    }

    if (diff < 604800) { // <1week: sun
        return date.toLocaleDateString('en-US', { weekday: 'short' });
    }

    if (diff < 31104000) { // <1year: 12 jan
        return date.toLocaleDateString('en-US', { day: 'numeric', month: 'short' });
    }

    // >1year: 2 year
    const years = Math.round(diff / 31104000);
    return `${years} year${years > 1 ? 's' : ''}`;
}
