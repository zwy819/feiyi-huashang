import { computed, ref, watch } from "vue";
import { Raycaster, Vector3, type Group, type PerspectiveCamera, type WebGLRenderer } from "three";
import type { Annotation } from "@/types/annotation";

interface HotspotScreen {
  annotation: Annotation;
  x: number;
  y: number;
  visible: boolean;
}

export function useModelHotspots(
  annotations: () => Annotation[],
  enabled: () => boolean,
  sceneState: () => {
    renderer: WebGLRenderer | null;
    camera: PerspectiveCamera | null;
    currentModel: Group | null;
  },
  onSelect: (annotation: Annotation) => void,
) {
  const showAnnotations = ref(true);
  const activeAnnotationId = ref<number | null>(null);
  const screens = ref<HotspotScreen[]>([]);
  const visibleHotspots = computed(() => showAnnotations.value ? screens.value : []);
  const hotspotsActive = computed(() => annotations().length > 0 && enabled());
  const point = new Vector3();
  const projected = new Vector3();
  const direction = new Vector3();
  const raycaster = new Raycaster();

  watch(annotations, list => {
    screens.value = list.map(annotation => ({ annotation, x: 0, y: 0, visible: false }));
    if (activeAnnotationId.value !== null && !list.some(a => a.id === activeAnnotationId.value)) {
      activeAnnotationId.value = null;
    }
  }, { immediate: true });

  function updateHotspots(): void {
    if (!hotspotsActive.value || !showAnnotations.value) return;
    const { renderer, camera, currentModel } = sceneState();
    if (!renderer || !camera || !currentModel) return;
    const { clientWidth: width, clientHeight: height } = renderer.domElement;
    for (const spot of screens.value) {
      point.set(spot.annotation.position_x, spot.annotation.position_y, spot.annotation.position_z);
      const distance = point.distanceTo(camera.position);
      projected.copy(point).project(camera);
      if (projected.z < -1 || projected.z > 1 || Math.abs(projected.x) > 1.1 || Math.abs(projected.y) > 1.1) {
        spot.visible = false;
        continue;
      }
      direction.copy(point).sub(camera.position).normalize();
      raycaster.set(camera.position, direction);
      raycaster.far = Math.max(distance - 0.05, 0.01);
      spot.visible = raycaster.intersectObject(currentModel, true).length === 0;
      spot.x = (projected.x * 0.5 + 0.5) * width;
      spot.y = (-projected.y * 0.5 + 0.5) * height;
    }
  }

  function onHotspotClick(annotation: Annotation): void {
    activeAnnotationId.value = annotation.id;
    onSelect(annotation);
  }

  return { showAnnotations, activeAnnotationId, visibleHotspots, hotspotsActive, updateHotspots, onHotspotClick };
}
