import type { LayoutServerLoad } from './$types'

export const load: LayoutServerLoad = async ({ cookies }) => {
	const accessToken = cookies.get('AccessToken');
	
	return {
		user: {
			loggedIn: !!accessToken
		}
	};
};