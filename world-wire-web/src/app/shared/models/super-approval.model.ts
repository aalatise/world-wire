import { Injectable, isDevMode } from '@angular/core';
import { ApprovalInfo, Approval } from './approval.interface';
import { environment } from '../../../environments/environment';
import { HttpClient } from '@angular/common/http';
import { PortalApiService } from '../services/portal-api.service';

@Injectable()
export class SuperApprovalsModel {

    constructor(
        private portalApiService: PortalApiService,
        private http: HttpClient,
    ) { }

    /**
     * Get super_approval info based on approval Id
     *
     * @param {string} approvalId
     * @returns {Promise<ApprovalInfo>}
     * @memberof SuperApprovalsModel
     */
    async getApprovalInfo(approvalId: string): Promise<ApprovalInfo> {

        try {

            const approval: Approval = await this.portalApiService.getSuperApprovalP(approvalId);
            const approvalInfo: ApprovalInfo = {
                key: approvalId,
            };

            // get user emails of uids
            const userPromises: Promise<any>[] = [];

            if (approval.uid_request) {
                const promise = this.portalApiService.getUserProfileP(approval.uid_request);
                userPromises.push(promise);
            }

            if (approval.uid_approve) {
              const promise = this.portalApiService.getUserProfileP(approval.uid_approve);
              userPromises.push(promise);
            }

            const users = await Promise.all(userPromises);

            approvalInfo.requestInitiatedBy = users.length > 0 ? users[0].profile.email : '';
            approvalInfo.requestApprovedBy = users.length > 1 ? users[1].profile.email : '';

            return approvalInfo;

        } catch (err) {
            if (isDevMode()) {
                console.log(err);
            }

            return Promise.reject('Approval not found');
        }

    }

    /**
     * Resets approval in case of error
     *
     * @param {string} approvalId
     * @returns {Promise<any>}
     * @memberof SuperApprovalsModel
     */
    async resetApprovals(approvalId: string): Promise<any> {

        const updateFields = {
            status: 'request',
            uid_approve: '',
        };

        // reset approval Id
        return await this.portalApiService.updateSuperApproval(approvalId, updateFields);
    }
}
