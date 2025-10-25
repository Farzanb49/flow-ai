import axios from axios;

export type Deployment = {
  id: string;
  project: string;
  namespace: string;
  image: string;
  status: string;
  description: string;
  createdAt: string;
};

const API_BASE: string = (import.meta as any).env?.VITE_API_BASE || http://localhost:8080;

export async function listDeployments(): Promise<Deployment[]> {
  const { data } = await axios.get(API_BASE + /deployments);
  return data;
}
