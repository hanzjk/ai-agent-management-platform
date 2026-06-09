/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { globalConfig, type Environment } from '@agent-management-platform/types';
import { Box, Typography, Button, Skeleton } from "@wso2/oxygen-ui";
import { Clock as AccessTime, Settings } from "@wso2/oxygen-ui-icons-react";
import { useParams, useSearchParams } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";
import {
  useGetAgent,
  useGetProject,
  useListDeploymentPipelines,
  useListEnvironments,
} from "@agent-management-platform/api-client";
import { EnvironmentCard } from "@agent-management-platform/shared-component";
import { InstrumentationDrawer } from "./InstrumentationDrawer";
import { NoDataFound } from "@agent-management-platform/views";
import { formatDistanceToNow } from "date-fns";

export const ExternalAgentOverview = () => {
  const { agentId, orgId, projectId } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const isInstrumentationDrawerOpen = searchParams.get("setup") === "true";
  const [selectedEnvironmentId, setSelectedEnvironmentId] =
    useState<string>("");

  const { data: agent } = useGetAgent({
    orgName: orgId,
    projName: projectId,
    agentName: agentId,
  });

  const { data: environmentList, isLoading: isEnvironmentsLoading } = useListEnvironments(
    { orgName: orgId },
  );
  const { data: project } = useGetProject({ orgName: orgId, projName: projectId });
  const { data: pipelinesData } = useListDeploymentPipelines({ orgName: orgId });

  const pipelineEnvOrder = useMemo(() => {
    const paths = pipelinesData?.deploymentPipelines
      ?.find((p) => p.name === project?.deploymentPipeline)?.promotionPaths ?? [];
    if (!paths.length) return [];
    const allTargets = new Set(
      paths.flatMap((p) => p.targetEnvironmentRefs.map((t) => t.name)),
    );
    const adjacency = new Map(
      paths.map((p) => [p.sourceEnvironmentRef, p.targetEnvironmentRefs.map((t) => t.name)]),
    );
    const roots = [...new Set(paths.map((p) => p.sourceEnvironmentRef))]
      .filter((s) => !allTargets.has(s));
    const chain: string[] = [];
    const visited = new Set<string>();
    let current: string | undefined = roots[0];
    while (current && !visited.has(current)) {
      chain.push(current);
      visited.add(current);
      current = (adjacency.get(current) ?? [])[0];
    }
    allTargets.forEach((t) => { if (!visited.has(t)) chain.push(t); });
    return chain;
  }, [pipelinesData, project?.deploymentPipeline]);

  const sortedEnvironmentList = useMemo(() => {
    if (!environmentList) return [];
    const orderIndex = new Map(pipelineEnvOrder.map((name, idx) => [name, idx]));
    return [...environmentList].sort((a, b) => {
      const ai = orderIndex.get(a.name);
      const bi = orderIndex.get(b.name);
      if (ai !== undefined && bi !== undefined) return ai - bi;
      if (ai !== undefined) return -1;
      if (bi !== undefined) return 1;
      return Number(b.isProduction) - Number(a.isProduction);
    });
  }, [environmentList, pipelineEnvOrder]);

  useEffect(() => {
    if (!selectedEnvironmentId && sortedEnvironmentList) {
      setSelectedEnvironmentId(sortedEnvironmentList?.[0]?.id ?? "");
    }
  }, [sortedEnvironmentList, selectedEnvironmentId]);

  const createdAtText = agent?.createdAt
    ? formatDistanceToNow(new Date(agent.createdAt), { addSuffix: true })
    : "—";

  const agentInstrumentationUrl = globalConfig.instrumentationUrl || "http://localhost:22893/otel";
  const apiKey = "ey***";

  const handleSetupAgent = (environmentId: string) => {
    setSelectedEnvironmentId(environmentId);
    setSearchParams({ setup: "true" });
  };

  return (
    <>
      <Box display="flex" flexDirection="column" gap={4}>
        <Box
          sx={{
            maxWidth: "fit-content",
            gap: 1.5,
            display: "flex",
            flexDirection: "column",
          }}
        >
          <Box display="flex" flexDirection="row" gap={1} alignItems="center">
            <Typography variant="body2">Created</Typography>
            <AccessTime size={14} />
            <Typography variant="body2">{createdAtText}</Typography>
          </Box>
        </Box>
        {isEnvironmentsLoading && (
          <Box display="flex" flexDirection="column" gap={2}>
            <Skeleton variant="rounded" height={100} />
            <Skeleton variant="rounded" height={100} />
          </Box>
        )}
        {!isEnvironmentsLoading &&
          sortedEnvironmentList &&
          sortedEnvironmentList.length > 0 && (
            <>
              {sortedEnvironmentList.map(
                (environment: Environment) =>
                  environment && (
                    <EnvironmentCard
                      key={environment.name}
                      orgId={orgId ?? "default"}
                      projectId={projectId ?? "default"}
                      agentId={agentId ?? "default"}
                      environment={environment}
                      actions={
                        <Button
                          variant="text"
                          size="small"
                          startIcon={<Settings size={16} />}
                          onClick={() =>
                            handleSetupAgent(environment.id ?? "")
                          }
                        >
                          Setup Agent
                        </Button>
                      }
                    />
                  )
              )}
            </>
          )}
        {!isEnvironmentsLoading &&
          (!sortedEnvironmentList || sortedEnvironmentList.length === 0) && (
            <NoDataFound
              message="No environments found"
              subtitle="Environments will appear here once they are created"
            />
          )}
      </Box>
      <InstrumentationDrawer
        open={isInstrumentationDrawerOpen}
        onClose={() => setSearchParams({})}
        agentId={agentId ?? ""}
        orgName={orgId ?? "default"}
        projName={projectId ?? "default"}
        agentName={agentId ?? ""}
        environment={
          sortedEnvironmentList?.find((env: Environment) =>
            env.id === selectedEnvironmentId)?.name
        }
        instrumentationUrl={agentInstrumentationUrl}
        apiKey={apiKey}
        componentUid={agent?.uuid}
        environmentUid={selectedEnvironmentId}
      />
    </>
  );
};
