import type { PageLoad } from './$types';

export const load: PageLoad = ({ params, url }) => {
  const parsedStatus = Number(params.status);
  const status = parsedStatus === 403 || parsedStatus === 404 ? parsedStatus : 500;

  return {
    status,
    from: url.searchParams.get('from')
  };
};
