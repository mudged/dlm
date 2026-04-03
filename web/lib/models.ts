export type ModelSummary = {
  id: string;
  name: string;
  created_at: string;
  light_count: number;
};

export type Light = { id: number; x: number; y: number; z: number };

export type ModelDetail = ModelSummary & { lights: Light[] };
