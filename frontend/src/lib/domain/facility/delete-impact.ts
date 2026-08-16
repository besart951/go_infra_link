export type FacilityDeleteImpactResource = 'apparat' | 'system_part';

export interface FacilityDeleteImpactBlocker {
  resource: string;
  count: number;
}

export interface FacilityDeleteImpact {
  resource: FacilityDeleteImpactResource;
  id: string;
  blockers: FacilityDeleteImpactBlocker[];
}

export interface FacilityDeleteImpactListResponse {
  items: FacilityDeleteImpact[];
}
