// Null Sims loads no analytics snippet: upstream's pages loaded one reporting to WoWSims' own
// property. These report only where a page defines `gtag`, and do nothing everywhere else.
const hasAnalytics = () => typeof gtag === 'function';

export const trackPageView = (title: string, slug: string) => {
	if (!hasAnalytics()) return;
	const normalizedSlug = slug.startsWith('/') ? slug.slice(1) : slug;
	gtag('event', 'page_view', {
		page_title: title,
		page_location: `${window.location.href}${normalizedSlug}`,
	});
};

export type TrackEventProps = {
	action: 'settings' | 'sim' | 'click';
	category: string;
	label?: string;
	value?: number | string | boolean;
	additionalData?: Record<string, string | number>;
};

export const trackEvent = ({ action, category, label, value, additionalData }: TrackEventProps) => {
	if (!hasAnalytics()) return;
	gtag('event', action, {
		event_category: category,
		event_label: label,
		event_value: typeof value !== 'undefined' ? String(value) : undefined,
		...additionalData,
	});
};
