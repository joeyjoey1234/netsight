import { create } from 'zustand';
import type { Project } from '../types';

interface ProjectState {
  currentProject: Project | null;
  projects: Project[];
  setCurrentProject: (project: Project | null) => void;
  setProjects: (projects: Project[]) => void;
  upsertProject: (project: Project) => void;
}

export const useProjectStore = create<ProjectState>((set) => ({
  currentProject: null,
  projects: [],
  setCurrentProject: (project) => set({ currentProject: project }),
  setProjects: (projects) => set({ projects }),
  upsertProject: (project) => set((state) => ({
    projects: state.projects.some((item) => item.id === project.id)
      ? state.projects.map((item) => item.id === project.id ? project : item)
      : [project, ...state.projects],
  })),
}));
