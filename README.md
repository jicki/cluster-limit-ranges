# cluster limit
    用于创建与管理 k8s resource `kind: LimitRange`


## LimitRange
    超出限制会在 `k8s events` 报错 如 `ephemeral storage`
       `Pod ephemeral local storage usage exceeds the total limit of containers 1Gi.`