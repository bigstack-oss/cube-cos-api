package triggers

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *helper) createConfigMapWithScript() error {
	config := h.genScriptConfigMap()
	h.kubernetes.SetConfigMapClient(triggers.DryRunNamespace)
	_, err := h.kubernetes.CreateConfigMap(&config)
	if err != nil {
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
		"dry run job %s failed with status %s(%s)",
		h.getScriptName(),
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

func (h *helper) getScriptName() string {
	return fmt.Sprintf("%s-script", h.reqId)
}

func (h *helper) getDryRunLogs() (string, error) {
	h.kubernetes.SetPodClient(triggers.DryRunNamespace)
	pods, err := h.kubernetes.ListPod(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", h.getScriptName()),
	})
	if err != nil {
		log.Errorf("triggers(%s): failed to list pods for dry run job(%v)", h.reqId, err)
		return "", err
	}

	if len(pods.Items) == 0 {
		log.Warnf("triggers(%s): no pods found for dry run job", h.reqId)
		return "", fmt.Errorf("no pods found for dry run job %s", h.getScriptName())
	}

	twoMiB := int64(2 * 1024 * 1024)
	logs, err := h.kubernetes.GetPodLog(
		pods.Items[0].Name,
		&corev1.PodLogOptions{Follow: false, LimitBytes: &twoMiB},
	)
	if err != nil {
		log.Errorf("triggers(%s): failed to get logs for dry run job pod(%v)", h.reqId, err)
		return "", err
	}

	defer logs.Close()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, logs)
	if err != nil {
		log.Errorf("triggers(%s): failed to read logs for dry run job pod(%v)", h.reqId, err)
		return "", err
	}

	return buf.String(), nil
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
