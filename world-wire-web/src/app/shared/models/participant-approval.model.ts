import { Injectable, isDevMode } from '@angular/core';
import { ApprovalInfo, Approval } from './approval.interface';
import { PortalApiService } from '../services/portal-api.service';

@Injectable()
export class ParticipantApprovalModel {

    constructor(
        private portalApiService: PortalApiService
    ) { }

    /**
     * Get participant_approval info based on approval Id
     *
     * @param {string} approvalId
     * @returns {Promise<ApprovalInfo>}
     * @memberof ParticipantApprovalModel
     */
    async getApprovalInfo(approvalId: string): Promise<ApprovalInfo> {

        if (isDevMode) {
            console.log('approvalId', approvalId);
        }

        const data = await this.portalApiService.getParticipantApprovalP(approvalId);

        const approval: Approval = data ? data : null;

        const approvalInfo: ApprovalInfo = {
          key: approvalId,
        };

        if (approval) {

          // get user emails of uids
          const userRequests: Promise<any>[] = [];

          if (approval.uid_request) {
            userRequests.push(this.portalApiService.getUserProfileP(approval.uid_request));
          }

          if (approval.uid_approve) {
            userRequests.push(this.portalApiService.getUserProfileP(approval.uid_approve));
          }

          const users = await Promise.all(userRequests);

          approvalInfo.requestInitiatedBy = users ? users[0].profile.email : null;
          approvalInfo.requestApprovedBy = users.length > 1 ? users[1].profile.email : null;
        }

        return approvalInfo;
    }

    /**
     * Resets approval in case of error
     *
     * @param {string} approvalId
     * @returns {Promise<any>}
     * @memberof ParticipantApprovalModel
     */
    async resetApprovals(approvalId: string): Promise<any> {
        const updateFields = {
            status: 'request',
            uid_approve: '',
        };

        return await this.portalApiService.updateParticipantApprovalP(approvalId, updateFields);
    }
}
