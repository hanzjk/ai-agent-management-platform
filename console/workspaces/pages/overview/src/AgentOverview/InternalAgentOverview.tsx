/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import {
  useGetAgent,
  useGetAgentBuilds,
  useGetProject,
  useListDeploymentPipelines,
  useListEnvironments,
} from "@agent-management-platform/api-client";
import {
  Clock as AccessTime,
  GitHub,
  CheckCircle,
} from "@wso2/oxygen-ui-icons-react";
import {
  Box,
  Button,
  CircularProgress,
  Typography,
  useTheme,
} from "@wso2/oxygen-ui";
import { generatePath, Link, useParams } from "react-router-dom";
import { useMemo } from "react";
import { formatDistanceToNow } from "date-fns";
import { EnvironmentCard } from "@agent-management-platform/shared-component";
import { absoluteRouteMap } from "@agent-management-platform/types";
import { KindInfoCard } from "./KindInfoCard";

export const InternalAgentOverview = () => {
  const { orgId, agentId, projectId } = useParams();
  const { data: agent } = useGetAgent({
    orgName: orgId,
    projName: projectId,
    agentName: agentId,
  });
  const { data: buildList } = useGetAgentBuilds({
    orgName: orgId,
    projName: projectId,
    agentName: agentId,
  });
  const { data: environmentList } = useListEnvironments({ orgName: orgId });
  const { data: project } = useGetProject({ orgName: orgId, projName: projectId });
  const { data: pipelinesData } = useListDeploymentPipelines({ orgName: orgId });
  const theme = useTheme();

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
    if (!environmentList || !pipelineEnvOrder.length) return [];
    const orderIndex = new Map(pipelineEnvOrder.map((name, idx) => [name, idx]));
    return environmentList
      .filter((e) => orderIndex.has(e.name))
      .sort((a, b) => (orderIndex.get(a.name) ?? 0) - (orderIndex.get(b.name) ?? 0));
  }, [environmentList, pipelineEnvOrder]);

  const isKindAgent = !!agent?.kindName;

  const createdAtText = agent?.createdAt
    ? formatDistanceToNow(new Date(agent.createdAt), { addSuffix: true })
    : "—";

  const repositoryUrl = useMemo(() => {
    const { appPath, branch, url } = agent?.provisioning?.repository ?? {};

    // If appPath is "/" (root), don't append it to avoid double slashes
    // Otherwise, remove the leading slash from appPath before appending
    if (appPath && appPath !== '/') {
      const normalizedPath = appPath.startsWith('/') ? appPath.substring(1) : appPath;
      return `${url}/tree/${branch}/${normalizedPath}`;
    }
    return `${url}/tree/${branch}`;
  }, [agent?.provisioning?.repository]);

  const loadingBuilds = useMemo(() => {
    return buildList?.builds.filter(
      (build) =>
        build.status === "Running" || build.status === "Pending"
    );
  }, [buildList]);

  return (
    <Box display="flex" flexDirection="column" gap={2}>
      <Box
        sx={{
          maxWidth: "fit-content",
          gap: 2,
          display: "flex",
          flexDirection: "column",
        }}
      >
        <Box display="flex" flexDirection="row" gap={1} alignItems="center">
          <Typography variant="body2">Created</Typography>
          <AccessTime size={14} />
          <Typography variant="body2">{createdAtText}</Typography>
        </Box>
        {isKindAgent ? (
          <KindInfoCard
            orgId={orgId ?? ""}
            kindName={agent!.kindName!}
          />
        ) : (
          <Box display="flex" flexDirection="row" gap={1} alignItems="center">
            <Typography variant="body2" width={100} noWrap>
              Source Code:
            </Typography>
            <Button
              component="a"
              startIcon={
                <GitHub size={16} color={theme.palette.text.secondary} />
              }
              variant="text"
              color="inherit"
              size="small"
              href={repositoryUrl}
              target="_blank"
              rel="noopener noreferrer"
            >
              <Typography variant="body2" noWrap>
                {repositoryUrl}
              </Typography>
            </Button>
          </Box>
        )}
        {!isKindAgent && (<Box display="flex" flexDirection="row" gap={1} alignItems="center">
          <Typography variant="body2" width={100} noWrap>
            Build Status:
          </Typography>
          {loadingBuilds?.length && loadingBuilds.length > 0 ? (
            <Button
              variant="text"
              size="small"
              color="inherit"
              component={Link}
              to={generatePath(
                absoluteRouteMap.children.org.children.projects.children.agents
                  .children.build.path,
                {
                  orgId,
                  projectId,
                  agentId,
                }
              )}
              startIcon={<CircularProgress size={14} color="inherit" />}
            >
              Build In Progress
            </Button>
          ) : (
            <Button
              variant="text"
              size="small"
              color="inherit"
              component={Link}
              to={generatePath(
                absoluteRouteMap.children.org.children.projects.children.agents
                  .children.build.path,
                {
                  orgId,
                  projectId,
                  agentId,
                }
              )}
              startIcon={<CheckCircle size={16} />}
            >
              Build Completed
            </Button>
          )}
        </Box>
        )}

      </Box>

      {sortedEnvironmentList && sortedEnvironmentList?.length > 0 && (
        <>
          {sortedEnvironmentList.map(
            (environment) =>
              environment && (
                <EnvironmentCard
                  key={environment.name}
                  orgId={orgId ?? "default"}
                  projectId={projectId ?? "default"}
                  agentId={agentId ?? "default"}
                  environment={environment}
                />
              )
          )}
        </>
      )}
    </Box>
  );
};
