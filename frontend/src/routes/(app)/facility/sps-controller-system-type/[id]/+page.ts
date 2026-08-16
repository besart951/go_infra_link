import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types.js';

export const load: PageLoad = ({ params, url }) => {
  throw redirect(308, `/facility/sps-controller-system-types/${params.id}${url.search}`);
};
