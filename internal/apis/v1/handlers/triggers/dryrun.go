package triggers

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *helper) getScriptName() string {
	return fmt.Sprintf("%s-script", h.reqId)
}

func (h *helper) createConfigMapWithScript() error {
	config := h.genScriptConfigMap()
	h.kubernetes.SetConfigMapClient(triggers.DryRunNamespace)
	_, err := h.kubernetes.CreateConfigMap(&config)
	if err != nil {
		log.Errorf("triggers(%s): failed to create configmap with script(%v)", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) dryRunScript() error {
	job := h.genJob()
	h.kubernetes.SetJobClient(triggers.DryRunNamespace)
	_, err := h.kubernetes.CreateJob(&job)
	if err != nil {
		log.Errorf("triggers(%s): failed to create dry run job(%v)", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) waitDryRunResult() (string, error) {
	result, err := h.waitingForJobCompletion()
	if err != nil {
		return "", err
	}

	log, err := h.getDryRunLogs()
	if err != nil {
		return "", err
	}

	if result == status.Completed {
		return log, nil
	}

	return "", fmt.Errorf(
		"test finished with status %s. reason: %s",
		result,
		log,
	)
}

func (h *helper) waitingForJobCompletion() (string, error) {
	interval := time.Second * 2
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	attemptsMax := 60
	for i := range attemptsMax {
		log.Infof("triggers(%s): waiting for dry run job result, attempt %d/%d", h.reqId, i+1, attemptsMax)
		<-ticker.C

		h.kubernetes.SetJobClient(triggers.DryRunNamespace)
		job, err := h.kubernetes.GetJob(h.getScriptName())
		if err != nil {
			log.Warnf("triggers(%s): failed to get dry run job(%v)", h.reqId, err)
			continue
		}

		if job.Status.Succeeded > 0 {
			log.Infof("triggers(%s): dry run job succeeded", h.reqId)
			return status.Completed, nil
		}

		if job.Status.Failed > 0 {
			log.Errorf("triggers(%s): dry run job failed", h.reqId)
			return status.Failed, nil
		}
	}

	return status.Unknown, fmt.Errorf(
		"dry run job %s did not finish within the expected time frame",
		h.getScriptName(),
	)
}

func (h *helper) genScriptConfigMap() corev1.ConfigMap {
	script := h.verifyScript["script"]
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.getScriptName(),
			Namespace: triggers.DryRunNamespace,
		},
		Data: map[string]string{
			"script.sh": script,
		},
	}
}

func (h *helper) genJob() batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.getScriptName(),
			Namespace: triggers.DryRunNamespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "script-runner",
							Image:   triggers.DryRunOciImage,
							Command: []string{"/bin/sh", "/scripts/script.sh"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "script-volume",
									MountPath: "/scripts",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "script-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: h.getScriptName(),
									},
									DefaultMode: func() *int32 {
										var mode int32 = 0777
										return &mode
									}(),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (h *helper) getDryRunLogs() (string, error) {
	pod, err := h.getDryRunPod()
	if err != nil {
		return "", err
	}

	logs, err := h.getPodLogs(pod)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(
		logs,
		"/scripts/script.sh: ",
		"",
	), nil
}

func (h *helper) deleteDryRunArtifacts() {
	h.kubernetes.SetConfigMapClient(triggers.DryRunNamespace)
	err := h.kubernetes.DeleteConfigMap(h.getScriptName())
	if err != nil {
		log.Warnf("triggers(%s): failed to delete configmap for dry run job(%v)", h.reqId, err)
	}

	h.kubernetes.SetJobClient(triggers.DryRunNamespace)
	err = h.kubernetes.DeleteJob(h.getScriptName())
	if err != nil {
		log.Warnf("triggers(%s): failed to delete job for dry run job(%v)", h.reqId, err)
	}
}

func (h *helper) getDryRunPod() (*corev1.Pod, error) {
	h.kubernetes.SetPodClient(triggers.DryRunNamespace)
	pods, err := h.kubernetes.ListPod(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", h.getScriptName()),
	})
	if err != nil {
		log.Errorf("triggers(%s): failed to list pods for dry run job(%v)", h.reqId, err)
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for dry run job %s", h.getScriptName())
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodPending {
			return &pod, nil
		}
	}

	return nil, fmt.Errorf(
		"no completed pods found for dry run job %s",
		h.getScriptName(),
	)
}

func (h *helper) getPodLogs(pod *corev1.Pod) (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	twoMiB := int64(2 * 1024 * 1024)

	req := h.kubernetes.GetLogs(
		pod.Name,
		&corev1.PodLogOptions{
			Follow:     false,
			LimitBytes: &twoMiB,
		},
	)

	logs, err := req.Stream(ctx)
	if err != nil {
		log.Errorf("triggers(%s): failed to get logs for dry run job pod(%v)", h.reqId, err)
		return "", err
	}

	defer logs.Close()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, logs)
	if err != nil {
		log.Errorf("triggers(%s): failed to read logs for dry run pod(%v)", h.reqId, err)
		return "", err
	}

	return buf.String(), nil
}
